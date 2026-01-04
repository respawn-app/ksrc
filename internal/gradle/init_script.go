package gradle

const initScript = `
import org.gradle.api.artifacts.component.ModuleComponentIdentifier
import org.gradle.api.artifacts.result.ResolvedArtifactResult
import org.gradle.jvm.JvmLibrary
import org.gradle.language.jvm.artifact.SourcesArtifact

fun splitCsv(value: String?): Set<String> {
    if (value == null) return emptySet()
    return value.split(",").map { it.trim() }.filter { it.isNotEmpty() }.toSet()
}

fun matchesGlob(patterns: String?, value: String): Boolean {
    if (patterns == null || patterns.isBlank()) return true
    return patterns.split(",").map { it.trim() }.filter { it.isNotEmpty() }.any { pattern ->
        val regex = pattern
            .replace(".", "\\.")
            .replace("*", ".*")
            .replace("?", ".")
            .toRegex()
        regex.matches(value)
    }
}

fun matchesModule(selector: String?, group: String, artifact: String, version: String): Boolean {
    if (selector == null || selector.isBlank()) return true
    if (selector.contains(":")) {
        val parts = selector.split(":")
        if (parts.size >= 2) {
            if (!matchesGlob(parts[0], group)) return false
            if (!matchesGlob(parts[1], artifact)) return false
            if (parts.size >= 3 && parts[2].isNotBlank()) {
                return matchesGlob(parts[2], version)
            }
            return true
        }
    }
    val candidates = listOf(
        group,
        artifact,
        "$group:$artifact",
        artifact.replace("-", "."),
        "$group:$artifact".replace("-", ".")
    )
    return candidates.any { matchesGlob(selector, it) || it.contains(selector) }
}

fun isSelectedProject(pathOrName: String, projectPath: String, projectName: String): Boolean {
    if (pathOrName.startsWith(":")) return pathOrName == projectPath
    return pathOrName == projectName
}

fun matchesTargets(name: String, targets: Set<String>): Boolean {
    if (targets.isEmpty()) return true
    if (name == "compileClasspath" || name == "runtimeClasspath" || name == "testCompileClasspath" || name == "testRuntimeClasspath") return true
    if (name.startsWith("common")) return true
    return targets.any { target -> name.startsWith(target) }
}

fun matchesScope(name: String, scope: String): Boolean {
    return when (scope) {
        "compile" -> name == "compileClasspath" || name.endsWith("MainCompileClasspath")
        "runtime" -> name == "runtimeClasspath" || name.endsWith("MainRuntimeClasspath")
        "test" -> name == "testCompileClasspath" || name == "testRuntimeClasspath" || name.endsWith("TestCompileClasspath") || name.endsWith("TestRuntimeClasspath")
        "all" -> true
        else -> name == "compileClasspath" || name.endsWith("MainCompileClasspath")
    }
}

gradle.projectsEvaluated {
    val props = gradle.startParameter.projectProperties
    val moduleProp = props["ksrcModule"] as String?
    val groupProp = props["ksrcGroup"] as String?
    val artifactProp = props["ksrcArtifact"] as String?
    val versionProp = props["ksrcVersion"] as String?
    val configProp = props["ksrcConfig"] as String?
    val targetsProp = props["ksrcTargets"] as String?
    val subprojectsProp = props["ksrcSubprojects"] as String?
    val scopeProp = (props["ksrcScope"] as String?) ?: "compile"
    val depProp = props["ksrcDep"] as String?

    val configs = splitCsv(configProp)
    val targets = splitCsv(targetsProp)
    val subprojects = splitCsv(subprojectsProp)

    val root = gradle.rootProject
    root.tasks.register("ksrcSources") {
        doLast {
            val projects = if (subprojects.isEmpty()) {
                gradle.allprojects
            } else {
                gradle.allprojects.filter { p -> subprojects.any { isSelectedProject(it, p.path, p.name) } }
            }

            val selectedConfigs = mutableListOf<org.gradle.api.artifacts.Configuration>()
            if (depProp == null) {
                projects.forEach { p ->
                    p.configurations.forEach { cfg ->
                        if (!cfg.isCanBeResolved) return@forEach
                        if (configs.isNotEmpty()) {
                            if (configs.contains(cfg.name)) {
                                selectedConfigs.add(cfg)
                            }
                        } else {
                            if (matchesTargets(cfg.name, targets) && matchesScope(cfg.name, scopeProp)) {
                                selectedConfigs.add(cfg)
                            }
                        }
                    }
                }
            }

            val moduleIds = linkedSetOf<ModuleComponentIdentifier>()
            if (depProp != null) {
                val dep = root.dependencies.create(depProp)
                val detached = root.configurations.detachedConfiguration(dep)
                detached.isTransitive = false
                detached.incoming.resolutionResult.allComponents.forEach { comp ->
                    val id = comp.id
                    if (id is ModuleComponentIdentifier) {
                        moduleIds.add(id)
                    }
                }
            } else {
                selectedConfigs.forEach { cfg ->
                    cfg.incoming.resolutionResult.allComponents.forEach { comp ->
                        val id = comp.id
                        if (id is ModuleComponentIdentifier) {
                            moduleIds.add(id)
                        }
                    }
                }
            }

            val filteredIds = moduleIds.filter { id ->
                matchesModule(moduleProp, id.group, id.module, id.version) &&
                matchesGlob(groupProp, id.group) &&
                matchesGlob(artifactProp, id.module) &&
                matchesGlob(versionProp, id.version)
            }

            filteredIds.forEach { id ->
                println("KSRCDEP|${'$'}{id.group}:${'$'}{id.module}:${'$'}{id.version}")
            }

            if (filteredIds.isEmpty()) return@doLast

            val query = root.dependencies.createArtifactResolutionQuery()
                .forComponents(filteredIds)
                .withArtifacts(JvmLibrary::class.java, SourcesArtifact::class.java)

            val result = query.execute()
            result.resolvedComponents.forEach { comp ->
                val id = comp.id
                if (id is ModuleComponentIdentifier) {
                    comp.getArtifacts(SourcesArtifact::class.java).forEach { art ->
                        if (art is ResolvedArtifactResult) {
                            println("KSRC|${'$'}{id.group}:${'$'}{id.module}:${'$'}{id.version}|${'$'}{art.file.absolutePath}")
                        }
                    }
                }
            }
        }
    }
}
`

func InitScript() string {
	return initScript
}
