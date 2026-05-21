import Foundation
import os.log

class StartupManager {
    static let shared = StartupManager()

    private let logger = Logger(subsystem: "com.neuro.ac", category: "Startup")
    private var startTime: CFAbsoluteTime = 0
    private var coldStart = true

    enum StartupPhase: String {
        case preInit = "pre_init"
        case sdkInit = "sdk_init"
        case uiLoad = "ui_load"
        case dataFetch = "data_fetch"
        case complete = "complete"
    }

    private var phaseTimestamps: [StartupPhase: CFAbsoluteTime] = [:]
    private var deferredTasks: [() -> Void] = []

    func markBegin() {
        startTime = CFAbsoluteTimeGetCurrent()
        phaseTimestamps[.preInit] = startTime
        logger.info("Startup began")
    }

    func markPhase(_ phase: StartupPhase) {
        let now = CFAbsoluteTimeGetCurrent()
        phaseTimestamps[phase] = now
        let elapsed = (now - startTime) * 1000
        logger.info("Startup phase \(phase.rawValue) at \(elapsed)ms")
    }

    func markComplete() {
        let now = CFAbsoluteTimeGetCurrent()
        phaseTimestamps[.complete] = now
        let totalDuration = (now - startTime) * 1000
        let type = coldStart ? "cold" : "warm"
        logger.info("Startup complete (\(type)): \(totalDuration)ms")
        coldStart = false
        runDeferredTasks()
    }

    func deferUntilReady(_ task: @escaping () -> Void) {
        if phaseTimestamps[.complete] != nil {
            task()
        } else {
            deferredTasks.append(task)
        }
    }

    private func runDeferredTasks() {
        let tasks = deferredTasks
        deferredTasks.removeAll()
        for task in tasks {
            task()
        }
    }

    func getMetrics() -> [String: Double] {
        var metrics: [String: Double] = [:]
        guard let start = phaseTimestamps[.preInit],
              let end = phaseTimestamps[.complete] else { return metrics }
        metrics["total_duration_ms"] = (end - start) * 1000
        metrics["is_cold_start"] = coldStart ? 1.0 : 0.0
        if let sdk = phaseTimestamps[.sdkInit] {
            metrics["sdk_init_ms"] = (sdk - start) * 1000
        }
        if let ui = phaseTimestamps[.uiLoad] {
            metrics["ui_load_ms"] = (ui - start) * 1000
        }
        if let data = phaseTimestamps[.dataFetch] {
            metrics["data_fetch_ms"] = (data - start) * 1000
        }
        return metrics
    }
}
