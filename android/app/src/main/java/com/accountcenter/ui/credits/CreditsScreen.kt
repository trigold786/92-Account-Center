package com.accountcenter.ui.credits

import android.content.ClipData
import android.content.ClipboardManager
import android.content.Context
import androidx.compose.foundation.background
import androidx.compose.foundation.layout.*
import androidx.compose.foundation.lazy.LazyColumn
import androidx.compose.foundation.shape.RoundedCornerShape
import androidx.compose.material.icons.Icons
import androidx.compose.material.icons.filled.ArrowBack
import androidx.compose.material.icons.filled.Refresh
import androidx.compose.material3.*
import androidx.compose.runtime.*
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.graphics.Brush
import androidx.compose.ui.graphics.Color
import androidx.compose.ui.platform.LocalContext
import androidx.compose.ui.text.TextStyle
import androidx.compose.ui.text.font.FontWeight
import androidx.compose.ui.unit.dp
import androidx.compose.ui.unit.sp
import androidx.hilt.navigation.compose.hiltViewModel
import androidx.lifecycle.compose.collectAsStateWithLifecycle
import androidx.compose.ui.res.stringResource
import com.accountcenter.R
import com.accountcenter.model.ReferralSummary
import com.accountcenter.model.Transaction
import com.accountcenter.ui.components.AppCard
import com.accountcenter.ui.theme.*

@OptIn(ExperimentalMaterial3Api::class)
@Composable
fun CreditsScreen(
    viewModel: CreditsViewModel = hiltViewModel(),
    onNavigateBack: () -> Unit
) {
    val balance by viewModel.balance.collectAsStateWithLifecycle()
    val transactions by viewModel.transactions.collectAsStateWithLifecycle()
    val referralSummary by viewModel.referralSummary.collectAsStateWithLifecycle()
    val isLoading by viewModel.isLoading.collectAsStateWithLifecycle()
    val hasMore by viewModel.hasMore.collectAsStateWithLifecycle()
    val context = LocalContext.current

    var referralLink by remember { mutableStateOf<String?>(null) }

    LaunchedEffect(referralLink) {
        referralLink?.let { link ->
            val clipboard = context.getSystemService(Context.CLIPBOARD_SERVICE) as ClipboardManager
            clipboard.setPrimaryClip(ClipData.newPlainText("referral", link))
            referralLink = null
        }
    }

    Box(modifier = Modifier.fillMaxSize().background(BgPrimary)) {
        Column(modifier = Modifier.fillMaxSize()) {
            CenterAlignedTopAppBar(
                title = { Text(stringResource(R.string.credits_title), color = TextPrimary) },
                navigationIcon = {
                    IconButton(onClick = onNavigateBack) {
                        Icon(Icons.Default.ArrowBack, contentDescription = stringResource(R.string.back), tint = TextPrimary)
                    }
                },
                actions = {
                    IconButton(onClick = { viewModel.refresh() }) {
                        Icon(Icons.Default.Refresh, contentDescription = stringResource(R.string.refresh), tint = TextPrimary)
                    }
                },
                colors = TopAppBarDefaults.centerAlignedTopAppBarColors(
                    containerColor = BgPrimary.copy(alpha = 0.9f)
                )
            )

            if (isLoading) {
                Box(modifier = Modifier.fillMaxSize(), contentAlignment = Alignment.Center) {
                    CircularProgressIndicator(color = BrandSecondary)
                }
            } else {
                LazyColumn(
                    modifier = Modifier.fillMaxSize().padding(16.dp),
                    verticalArrangement = Arrangement.spacedBy(12.dp)
                ) {
                    item {
                        BalanceCard(balance)
                    }

                    referralSummary?.let { summary ->
                        item {
                            ReferralCard(summary) {
                                viewModel.generateReferralLink { link -> referralLink = link }
                            }
                        }
                    }

                    item {
                        Spacer(modifier = Modifier.height(8.dp))
                        Text(stringResource(R.string.credits_transaction_history), style = MaterialTheme.typography.labelSmall, color = TextSecondary)
                    }

                    transactions.forEach { tx ->
                        item {
                            TransactionItem(tx)
                        }
                    }

                    if (hasMore) {
                        item {
                            TextButton(onClick = { viewModel.loadMore() }, modifier = Modifier.fillMaxWidth()) {
                                Text(stringResource(R.string.credits_load_more), color = BrandSecondary)
                            }
                        }
                    }
                }
            }
        }
    }
}

@Composable
private fun BalanceCard(balance: Double) {
    AppCard(
        modifier = Modifier.fillMaxWidth()
    ) {
        Column(
            modifier = Modifier.fillMaxWidth(),
            horizontalAlignment = Alignment.CenterHorizontally
        ) {
            Text(stringResource(R.string.credits_current), style = MaterialTheme.typography.bodyMedium, color = TextSecondary)
            Spacer(modifier = Modifier.height(8.dp))
            Text(
                text = "%.2f".format(balance),
                fontSize = 36.sp,
                fontWeight = FontWeight.Bold,
                style = TextStyle(brush = Brush.horizontalGradient(listOf(BrandPrimary, BrandSecondary)))
            )
        }
    }
}

@Composable
private fun ReferralCard(summary: ReferralSummary, onShare: () -> Unit) {
    AppCard(
        modifier = Modifier.fillMaxWidth()
    ) {
        Text(stringResource(R.string.credits_referral_promo), style = MaterialTheme.typography.titleSmall, fontWeight = FontWeight.Bold, color = TextPrimary)
        Spacer(modifier = Modifier.height(8.dp))
        Text("${stringResource(R.string.credits_referral_count)}：${summary.totalReferees}", style = MaterialTheme.typography.bodySmall, color = TextSecondary)
        Text("${stringResource(R.string.credits_active_friends)}：${summary.activeReferees}", style = MaterialTheme.typography.bodySmall, color = TextSecondary)
        Text("${stringResource(R.string.credits_total_earned)}：${summary.totalEarned} 积分", style = MaterialTheme.typography.bodySmall, color = TextSecondary)
        Spacer(modifier = Modifier.height(8.dp))
        Button(
            onClick = onShare,
            modifier = Modifier.fillMaxWidth(),
            colors = ButtonDefaults.buttonColors(containerColor = BrandPrimary),
            shape = RoundedCornerShape(12.dp)
        ) {
            Text(stringResource(R.string.credits_copy_referral), color = Color.White)
        }
    }
}

@Composable
private fun TransactionItem(transaction: Transaction) {
    Card(
        modifier = Modifier.fillMaxWidth(),
        colors = CardDefaults.cardColors(containerColor = BgCard),
        shape = RoundedCornerShape(16.dp)
    ) {
        Row(
            modifier = Modifier.padding(16.dp).fillMaxWidth(),
            verticalAlignment = Alignment.CenterVertically,
            horizontalArrangement = Arrangement.SpaceBetween
        ) {
            Column(modifier = Modifier.weight(1f)) {
                Text(transaction.details ?: transaction.type, style = MaterialTheme.typography.bodyMedium, color = TextPrimary)
                Text(transaction.createdAt, style = MaterialTheme.typography.bodySmall, color = TextSecondary)
            }
            val isPositive = transaction.type in listOf("earn", "referral_bonus", "refund")
            Text(
                if (isPositive) "+%.2f".format(transaction.amount) else "-%.2f".format(transaction.amount),
                color = if (isPositive) Success else Danger,
                fontWeight = FontWeight.Bold
            )
        }
    }
}
