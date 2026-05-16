package com.accountcenter

import com.accountcenter.model.ApiDataResponse
import com.accountcenter.model.CreditAccount
import com.accountcenter.model.ReferralSummary
import com.accountcenter.model.TransactionList
import com.google.gson.Gson
import com.google.gson.GsonBuilder
import com.google.gson.reflect.TypeToken
import org.junit.Assert.assertEquals
import org.junit.Assert.assertNotNull
import org.junit.Before
import org.junit.Test

class CreditsViewModelTest {

    private lateinit var gson: Gson

    @Before
    fun setUp() {
        gson = GsonBuilder().create()
    }

    @Test
    fun testDecodeCreditAccount() {
        val json = """
        {
            "code": 200,
            "message": "success",
            "data": {
                "user_id": 42,
                "balance": 1500.50,
                "status": "active"
            }
        }
        """.trimIndent()
        val type = object : TypeToken<ApiDataResponse<CreditAccount>>() {}.type
        val response: ApiDataResponse<CreditAccount> = gson.fromJson(json, type)
        assertNotNull(response)
        assertEquals(200, response.code)
        assertNotNull(response.data)
        assertEquals(1500.50, response.data!!.balance, 0.001)
        assertEquals("active", response.data!!.status)
    }

    @Test
    fun testDecodeTransactionList() {
        val json = """
        {
            "code": 200,
            "message": "success",
            "data": {
                "transactions": [
                    {"id":1,"credit_account_id":10,"type":"earn","amount":100.0,"reference_id":"ref_1","details":"邀请奖励","status":"completed","created_at":"2026-05-01T10:00:00Z"},
                    {"id":2,"credit_account_id":10,"type":"consume","amount":50.0,"reference_id":"ref_2","details":"积分消费","status":"completed","created_at":"2026-05-02T10:00:00Z"}
                ],
                "total": 2,
                "page": 1,
                "page_size": 20
            }
        }
        """.trimIndent()
        val type = object : TypeToken<ApiDataResponse<TransactionList>>() {}.type
        val response: ApiDataResponse<TransactionList> = gson.fromJson(json, type)
        assertNotNull(response)
        assertNotNull(response.data)
        assertEquals(2, response.data!!.transactions.size)
        assertEquals(2, response.data!!.total)
        assertEquals("earn", response.data!!.transactions[0].type)
        assertEquals(50.0, response.data!!.transactions[1].amount, 0.001)
    }

    @Test
    fun testDecodeReferralSummary() {
        val json = """
        {
            "code": 200,
            "message": "success",
            "data": {
                "total_referees": 10,
                "total_earned": 500.0,
                "active_referees": 5
            }
        }
        """.trimIndent()
        val type = object : TypeToken<ApiDataResponse<ReferralSummary>>() {}.type
        val response: ApiDataResponse<ReferralSummary> = gson.fromJson(json, type)
        assertNotNull(response)
        assertNotNull(response.data)
        assertEquals(10, response.data!!.totalReferees)
        assertEquals(500.0, response.data!!.totalEarned, 0.001)
        assertEquals(5, response.data!!.activeReferees)
    }
}
