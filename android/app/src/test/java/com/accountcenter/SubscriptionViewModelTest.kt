package com.accountcenter

import com.accountcenter.model.Subscription
import com.accountcenter.model.TierInfo
import com.google.gson.Gson
import com.google.gson.GsonBuilder
import com.google.gson.annotations.SerializedName
import org.junit.Assert.assertEquals
import org.junit.Assert.assertNotNull
import org.junit.Before
import org.junit.Test

class SubscriptionViewModelTest {

    private lateinit var gson: Gson

    @Before
    fun setUp() {
        gson = GsonBuilder().create()
    }

    @Test
    fun testDecodeSubscription() {
        val json = """
        {
            "id": 1,
            "user_id": 42,
            "tier_level": 2,
            "start_time": "2026-01-01T00:00:00Z",
            "end_time": "2026-12-31T23:59:59Z",
            "status": "active",
            "price": 99.99,
            "payment_method": "alipay",
            "order_id": "ORD_123"
        }
        """.trimIndent()
        val sub = gson.fromJson(json, Subscription::class.java)
        assertNotNull(sub)
        assertEquals(1L, sub.id)
        assertEquals(42L, sub.userId)
        assertEquals(2, sub.tierLevel)
        assertEquals("active", sub.status)
        assertEquals(99.99, sub.price, 0.001)
    }

    @Test
    fun testDecodeSubscriptionList() {
        val json = """
        [
            {"id":1,"user_id":42,"tier_level":1,"start_time":"2026-01-01T00:00:00Z","end_time":"2026-06-30T23:59:59Z","status":"expired","price":29.99},
            {"id":2,"user_id":42,"tier_level":2,"start_time":"2026-07-01T00:00:00Z","end_time":"2026-12-31T23:59:59Z","status":"active","price":99.99}
        ]
        """.trimIndent()
        val subs = gson.fromJson(json, Array<Subscription>::class.java)
        assertNotNull(subs)
        assertEquals(2, subs.size)
        assertEquals("expired", subs[0].status)
        assertEquals("active", subs[1].status)
    }

    @Test
    fun testDecodeTierInfo() {
        val json = """
        {
            "user_id": 42,
            "identity_tier": 2
        }
        """.trimIndent()
        val tier = gson.fromJson(json, TierInfo::class.java)
        assertNotNull(tier)
        assertEquals(42L, tier.userId)
        assertEquals(2, tier.identityTier)
    }

    @Test
    fun testActiveSubscriptionFilter() {
        val json = """
        [
            {"id":1,"user_id":42,"tier_level":1,"start_time":"2026-01-01T00:00:00Z","end_time":"2026-06-30T23:59:59Z","status":"expired","price":29.99},
            {"id":2,"user_id":42,"tier_level":2,"start_time":"2026-07-01T00:00:00Z","end_time":"2026-12-31T23:59:59Z","status":"active","price":99.99}
        ]
        """.trimIndent()
        val subs = gson.fromJson(json, Array<Subscription>::class.java).toList()
        val active = subs.filter { it.status == "active" }
        assertEquals(1, active.size)
        assertEquals(2, active.first().tierLevel)
    }
}
