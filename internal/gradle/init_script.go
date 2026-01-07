package gradle

const initScript = `
import org.gradle.api.artifacts.component.ModuleComponentIdentifier

def splitCsv = { String value ->
    if (value == null) return [] as Set
    value.split(',').collect { it.trim() }.findAll { it }.toSet()
}

def matchesGlob = { String patterns, String value ->
    if (patterns == null || patterns.trim().isEmpty()) return true
    patterns.split(',').collect { it.trim() }.findAll { it }.any { pattern ->
        def regex = pattern.replace('.', '\\.').replace('*', '.*').replace('?', '.')
        return value ==~ regex
    }
}

def matchesModule = { String selector, String group, String artifact, String version ->
    if (selector == null || selector.trim().isEmpty()) return true
    if (selector.contains(':')) {
        def parts = selector.split(':')
        if (parts.length >= 2) {
            if (!matchesGlob(parts[0], group)) return false
            if (!matchesGlob(parts[1], artifact)) return false
            if (parts.length >= 3 && parts[2]) return matchesGlob(parts[2], version)
            return true
        }
    }
    def candidates = [
        group,
        artifact,
        "${group}:${artifact}",
        artifact.replace('-', '.'),
        "${group}:${artifact}".replace('-', '.')
    ]
    return candidates.any { matchesGlob(selector, it) || it.contains(selector) }
}

def isSelectedProject = { String pathOrName, String projectPath, String projectName ->
    if (pathOrName.startsWith(':')) return pathOrName == projectPath
    return pathOrName == projectName
}

def matchesTargets = { String name, Set targets ->
    if (targets.isEmpty()) return true
    if (['compileClasspath','runtimeClasspath','testCompileClasspath','testRuntimeClasspath'].contains(name)) return true
    if (name.startsWith('common')) return true
    return targets.any { target -> name.startsWith(target) }
}

def matchesScope = { String name, String scope ->
    switch (scope) {
        case 'compile':
            return name == 'compileClasspath' || name.endsWith('MainCompileClasspath')
        case 'runtime':
            return name == 'runtimeClasspath' || name.endsWith('MainRuntimeClasspath')
        case 'test':
            return name == 'testCompileClasspath' || name == 'testRuntimeClasspath' || name.endsWith('TestCompileClasspath') || name.endsWith('TestRuntimeClasspath')
        case 'all':
            return true
        default:
            return name == 'compileClasspath' || name.endsWith('MainCompileClasspath')
    }
}

def props = gradle.startParameter.projectProperties
def moduleProp = props['ksrcModule']
def groupProp = props['ksrcGroup']
def artifactProp = props['ksrcArtifact']
def versionProp = props['ksrcVersion']
def configProp = props['ksrcConfig']
def targetsProp = props['ksrcTargets']
def subprojectsProp = props['ksrcSubprojects']
def scopeProp = props['ksrcScope'] ?: 'compile'
def depProp = props['ksrcDep']
def includeBuildscript = (props['ksrcBuildscript'] ?: 'true').toString().toBoolean()
def includeIncludedBuilds = (props['ksrcIncludeBuilds'] ?: 'true').toString().toBoolean()

gradle.rootProject { root ->

    def configs = splitCsv(configProp as String)
    def targets = splitCsv(targetsProp as String)
    def subprojects = splitCsv(subprojectsProp as String)

    def selectedProjects = subprojects.isEmpty() ? root.allprojects : root.allprojects.findAll { p ->
        subprojects.any { isSelectedProject(it as String, p.path, p.name) }
    }

    def emitForProject = { proj ->
        def selectedConfigs = []
        if (!depProp) {
            proj.configurations.each { cfg ->
                if (!cfg.canBeResolved) return
                if (!configs.isEmpty()) {
                    if (configs.any { cfgPattern -> matchesGlob(cfgPattern as String, cfg.name) }) selectedConfigs << cfg
                } else {
                    if (matchesTargets(cfg.name, targets) && matchesScope(cfg.name, scopeProp as String)) selectedConfigs << cfg
                }
            }
        } else {
            def dep = proj.dependencies.create(depProp as String)
            def detached = proj.configurations.detachedConfiguration(dep)
            detached.transitive = false
            selectedConfigs << detached
        }

        def moduleIds = [] as Set
        selectedConfigs.each { cfg ->
            cfg.incoming.resolutionResult.allComponents.each { comp ->
                def id = comp.id
                if (id instanceof ModuleComponentIdentifier) moduleIds << id
            }
        }

        if (includeBuildscript) {
            def buildscriptConfigs = []
            proj.buildscript.configurations.each { cfg ->
                if (!cfg.canBeResolved) return
                if (!configs.isEmpty()) {
                    if (configs.any { cfgPattern -> matchesGlob(cfgPattern as String, cfg.name) }) buildscriptConfigs << cfg
                } else {
                    buildscriptConfigs << cfg
                }
            }
            buildscriptConfigs.each { cfg ->
                cfg.incoming.resolutionResult.allComponents.each { comp ->
                    def id = comp.id
                    if (id instanceof ModuleComponentIdentifier) moduleIds << id
                }
            }
        }

        def filteredIds = moduleIds.findAll { id ->
            matchesModule(moduleProp as String, id.group, id.module, id.version) &&
            matchesGlob(groupProp as String, id.group) &&
            matchesGlob(artifactProp as String, id.module) &&
            matchesGlob(versionProp as String, id.version)
        }

        filteredIds.each { id ->
            println "KSRCDEP|${id.group}:${id.module}:${id.version}"
        }

        if (filteredIds.isEmpty()) return

        def sourceDeps = filteredIds.collect { id ->
            proj.dependencies.create([group: id.group, name: id.module, version: id.version, classifier: 'sources'])
        }
        if (sourceDeps.isEmpty()) return

        def sourcesCfg = proj.configurations.detachedConfiguration()
        sourceDeps.each { dep ->
            sourcesCfg.dependencies.add(dep)
        }
        sourcesCfg.transitive = false
        def lenient = sourcesCfg.resolvedConfiguration.lenientConfiguration
        lenient.artifacts.each { art ->
            def id = art.moduleVersion.id
            println "KSRC|${id.group}:${id.name}:${id.version}|${art.file.absolutePath}"
        }
    }

    if (depProp) {
        root.tasks.register('ksrcSources') {
            doLast {
                emitForProject(project)
            }
        }
        return
    }

    def projectTasks = []
    selectedProjects.each { p ->
        def projectRef = p
        projectRef.tasks.register('ksrcSourcesProject') {
            doLast {
                emitForProject(projectRef)
            }
        }
        projectTasks << projectRef.tasks.named('ksrcSourcesProject')
    }

    root.tasks.register('ksrcSources') {
        dependsOn(projectTasks)
    }
}

gradle.settingsEvaluated { settings ->
    if (!includeIncludedBuilds) return
    try {
        gradle.includedBuilds.each { build ->
            def dir = null
            try {
                dir = build.projectDir
            } catch (Throwable ignored) {
                dir = null
            }
            if (dir == null) {
                try {
                    dir = build.rootDir
                } catch (Throwable ignored) {
                    dir = null
                }
            }
            if (dir != null) {
                println "KSRCINCLUDE|${dir.absolutePath}"
            }
        }
    } catch (Throwable ignored) {
        // Included builds may be unavailable for this Gradle version or lifecycle phase.
    }
}
`

func InitScript() string {
	return initScript
}
