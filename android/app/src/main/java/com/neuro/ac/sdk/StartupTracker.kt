package com.neuro.ac.sdk

import android.os.Build
import android.os.Bundle
import android.util.Log
import java.util.concurrent.CopyOnWriteArrayList

class StartupTracker {
    companion object {
        private const val TAG = "StartupTracker"
        private const val PREFS_NAME = "startup_metrics"
        private const val KEY_LAST_COLD_START = "last_cold_start_ms"
        private const val KEY_LAST_WARM_START = "last_warm_start_ms"
    }

    enum class Phase(val key: String) {
        PRE_INIT("pre_init"),
        SDK_INIT("sdk_init"),
        UI_LOAD("ui_load"),
        DATA_FETCH("data_fetch"),
        COMPLETE("complete")
    }

    private val phaseTimestamps = mutableMapOf<Phase, Long>()
    private val deferredTasks = CopyOnWriteArrayList<() -> Unit>()
    private var startTime: Long = 0
    private var isColdStart: Boolean = true

    fun markBegin() {
        startTime = System.nanoTime()
        phaseTimestamps[Phase.PRE_INIT] = startTime
        Log.i(TAG, "Startup began")
    }

    fun markPhase(phase: Phase) {
        val now = System.nanoTime()
        phaseTimestamps[phase] = now
        val elapsedMs = (now - startTime) / 1_000_000
        Log.i(TAG, "Startup phase ${phase.key} at ${elapsedMs}ms")
    }

    fun markComplete(context: android.content.Context) {
        val now = System.nanoTime()
        phaseTimestamps[Phase.COMPLETE] = now
        val totalMs = (now - startTime) / 1_000_000
        val type = if (isColdStart) "cold" else "warm"
        Log.i(TAG, "Startup complete ($type): ${totalMs}ms")

        val prefs = context.getSharedPreferences(PREFS_NAME, android.content.Context.MODE_PRIVATE)
        prefs.edit()
            .putLong(if (isColdStart) KEY_LAST_COLD_START else KEY_LAST_WARM_START, totalMs)
            .apply()

        isColdStart = false
        runDeferredTasks()
    }

    fun deferUntilReady(task: () -> Unit) {
        if (phaseTimestamps.containsKey(Phase.COMPLETE)) {
            task()
        } else {
            deferredTasks.add(task)
        }
    }

    private fun runDeferredTasks() {
        val tasks = deferredTasks.toList()
        deferredTasks.clear()
        tasks.forEach { it() }
    }

    fun getMetrics(): Map<String, Long> {
        val metrics = mutableMapOf<String, Long>()
        val start = phaseTimestamps[Phase.PRE_INIT] ?: return metrics
        val end = phaseTimestamps[Phase.COMPLETE] ?: return metrics
        metrics["total_duration_ms"] = (end - start) / 1_000_000
        metrics["is_cold_start"] = if (isColdStart) 1L else 0L
        phaseTimestamps[Phase.SDK_INIT]?.let { metrics["sdk_init_ms"] = (it - start) / 1_000_000 }
        phaseTimestamps[Phase.UI_LOAD]?.let { metrics["ui_load_ms"] = (it - start) / 1_000_000 }
        phaseTimestamps[Phase.DATA_FETCH]?.let { metrics["data_fetch_ms"] = (it - start) / 1_000_000 }
        return metrics
    }

    fun getBaselineProfileHint(): String {
        return when {
            Build.VERSION.SDK_INT >= Build.VERSION_CODES.UPSIDE_DOWN_CAKE -> "BaselineProfile-1.0"
            Build.VERSION.SDK_INT >= Build.VERSION_CODES.S -> "StartupProfile-1.0"
            else -> ""
        }
    }
}
