package com.neuro.ac.sdk

import android.content.Context
import android.content.res.Configuration
import android.view.View
import android.view.ViewTreeObserver
import android.widget.Button
import android.widget.ImageView
import android.widget.TextView
import kotlin.math.max

object AccessibilityHelper {

    fun setContentDescription(view: View, description: String) {
        view.contentDescription = description
    }

    fun setContentDescription(imageView: ImageView, description: String) {
        imageView.contentDescription = description
    }

    fun ensureMinTouchTarget(view: View, minDp: Int = 48) {
        val density = view.context.resources.displayMetrics.density
        val minPx = (minDp * density).toInt()
        view.viewTreeObserver.addOnGlobalLayoutListener(object : ViewTreeObserver.OnGlobalLayoutListener {
            override fun onGlobalLayout() {
                view.viewTreeObserver.removeOnGlobalLayoutListener(this)
                val width = view.width
                val height = view.height
                if (width < minPx || height < minPx) {
                    val newWidth = max(width, minPx)
                    val newHeight = max(height, minPx)
                    val layoutParams = view.layoutParams
                    layoutParams.width = newWidth
                    layoutParams.height = newHeight
                    view.layoutParams = layoutParams
                }
            }
        })
    }

    fun applyFontScale(textView: TextView, scaleFactor: Float) {
        val config = Configuration(textView.context.resources.configuration)
        config.fontScale = scaleFactor
        val scaledContext = textView.context.createConfigurationContext(config)
        textView.textSize = textView.textSize * scaleFactor
    }

    fun isReduceMotionEnabled(context: Context): Boolean {
        val animatorDuration = android.animation.ValueAnimator.getDurationScale()
        return animatorDuration == 0f
    }

    fun setupAccessibleButton(button: Button, label: String, minTouchDp: Int = 48) {
        button.contentDescription = label
        button.importantForAccessibility = View.IMPORTANT_FOR_ACCESSIBILITY_YES
        ensureMinTouchTarget(button, minTouchDp)
    }

    fun announceForAccessibility(view: View, message: String) {
        view.announceForAccessibility(message)
    }

    fun setHeading(textView: TextView) {
        textView.setAccessibilityDelegate(object : View.AccessibilityDelegate() {
            override fun onInitializeAccessibilityNodeInfo(host: View, info: android.view.accessibility.AccessibilityNodeInfo) {
                super.onInitializeAccessibilityNodeInfo(host, info)
                info.heading = true
            }
        })
    }
}
