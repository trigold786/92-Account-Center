package com.accountcenter.ui.subscription

import androidx.compose.foundation.background
import androidx.compose.foundation.layout.*
import androidx.compose.foundation.lazy.LazyColumn
import androidx.compose.foundation.shape.RoundedCornerShape
import androidx.compose.material.icons.Icons
import androidx.compose.material.icons.filled.ArrowBack
import androidx.compose.material.icons.filled.Refresh
import androidx.compose.material3.*
import androidx.compose.runtime.Composable
import androidx.compose.runtime.getValue
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.draw.clip
import androidx.compose.ui.graphics.Color
import androidx.compose.ui.res.stringResource
import androidx.compose.ui.text.font.FontWeight
import androidx.compose.ui.unit.dp
import androidx.hilt.navigation.compose.hiltViewModel
import androidx.lifecycle.compose.collectAsStateWithLifecycle
import com.accountcenter.R
import com.accountcenter.model.Subscription
import com.accountcenter.model.TierInfo
import com.accountcenter.ui.components.AppCard
import com.accountcenter.ui.theme.*

@OptIn(ExperimentalMaterial3Api::class)
@Composable
fun SubscriptionScreen(
    viewModel: SubscriptionViewModel = hiltViewModel(),
    onNavigateBack: () -> Unit
) {
    val subscriptions by viewModel.subscriptions.collectAsStateWithLifecycle()
    val tier by viewModel.tier.collectAsStateWithLifecycle()
    val isLoading by viewModel.isLoading.collectAsStateWithLifecycle()

    Box(modifier = Modifier.fillMaxSize().background(BgPrimary)) {
        Column(modifier = Modifier.fillMaxSize()) {
            CenterAlignedTopAppBar(
                title = { Text(stringResource(R.string.subscription_title), color = TextPrimary) },
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
                    tier?.let { tierInfo ->
                        item {
                            TierBadge(tierInfo.identityTier)
                        }
                    }

                    val activeSubs = subscriptions.filter { it.status == "active" }
                    if (activeSubs.isNotEmpty()) {
                        item {
                            Text(stringResource(R.string.subscription_current), style = MaterialTheme.typography.labelSmall, color = TextSecondary)
                        }
                        activeSubs.forEach { sub ->
                            item {
                                ActiveSubscriptionCard(sub)
                            }
                        }
                    } else {
                        item {
                            Card(
                                modifier = Modifier.fillMaxWidth(),
                                colors = CardDefaults.cardColors(containerColor = BgCard),
                                shape = RoundedCornerShape(16.dp)
                            ) {
                                Text(
                                    stringResource(R.string.subscription_no_active),
                                    modifier = Modifier.padding(24.dp),
                                    style = MaterialTheme.typography.bodyLarge,
                                    color = TextSecondary
                                )
                            }
                        }
                    }

                    item {
                        Spacer(modifier = Modifier.height(8.dp))
                        Text(stringResource(R.string.subscription_history), style = MaterialTheme.typography.labelSmall, color = TextSecondary)
                    }
                    subscriptions.filter { it.status != "active" }.forEach { sub ->
                        item {
                            SubscriptionHistoryCard(sub)
                        }
                    }
                }
            }
        }
    }
}

@Composable
private fun TierBadge(level: Int) {
    val (name, tierColor) = when (level) {
        0 -> stringResource(R.string.tier_free) to TierFree
        1 -> stringResource(R.string.tier_basic) to TierBasic
        2 -> stringResource(R.string.tier_premium) to TierPremium
        3 -> stringResource(R.string.tier_enterprise) to TierEnterprise
        else -> stringResource(R.string.tier_free) to TierFree
    }
    AppCard(
        modifier = Modifier.fillMaxWidth()
    ) {
        Row(verticalAlignment = Alignment.CenterVertically) {
            Box(modifier = Modifier.size(12.dp).clip(RoundedCornerShape(6.dp)).background(tierColor))
            Spacer(modifier = Modifier.width(8.dp))
            Text(
                stringResource(R.string.subscription_tier),
                style = MaterialTheme.typography.bodySmall,
                color = TextSecondary
            )
            Spacer(modifier = Modifier.width(8.dp))
            Surface(
                color = BrandSecondary.copy(alpha = 0.15f),
                shape = RoundedCornerShape(8.dp)
            ) {
                Text(
                    "Lv.$level",
                    modifier = Modifier.padding(horizontal = 8.dp, vertical = 2.dp),
                    style = MaterialTheme.typography.labelSmall,
                    color = BrandSecondary
                )
            }
            Spacer(modifier = Modifier.width(8.dp))
            Text(
                name,
                style = MaterialTheme.typography.titleMedium,
                color = tierColor,
                fontWeight = FontWeight.Bold
            )
        }
    }
}

@Composable
private fun ActiveSubscriptionCard(subscription: Subscription) {
    Box(
        modifier = Modifier
            .fillMaxWidth()
            .clip(RoundedCornerShape(16.dp))
            .background(BgCard)
    ) {
        Row(modifier = Modifier.fillMaxWidth()) {
            Box(
                modifier = Modifier
                    .width(3.dp)
                    .fillMaxHeight()
                    .background(BrandPrimary)
            )
            Column(modifier = Modifier.padding(16.dp).weight(1f)) {
                Row {
                    Text("${stringResource(R.string.subscription_status)}：", style = MaterialTheme.typography.bodySmall, color = TextSecondary)
                    Text(subscription.status, style = MaterialTheme.typography.bodySmall, color = Success)
                }
                Spacer(modifier = Modifier.height(4.dp))
                Row {
                    Text("${stringResource(R.string.subscription_start_time)}：", style = MaterialTheme.typography.bodySmall, color = TextSecondary)
                    Text(subscription.startTime, style = MaterialTheme.typography.bodySmall, color = TextPrimary)
                }
                Row {
                    Text("${stringResource(R.string.subscription_end_time)}：", style = MaterialTheme.typography.bodySmall, color = TextSecondary)
                    Text(subscription.endTime, style = MaterialTheme.typography.bodySmall, color = TextPrimary)
                }
                Row {
                    Text("${stringResource(R.string.subscription_price)}：", style = MaterialTheme.typography.bodySmall, color = TextSecondary)
                    Text("¥${subscription.price}", style = MaterialTheme.typography.bodySmall, color = TextPrimary, fontWeight = FontWeight.Bold)
                }
            }
        }
    }
}

@Composable
private fun SubscriptionHistoryCard(subscription: Subscription) {
    Card(
        modifier = Modifier.fillMaxWidth(),
        colors = CardDefaults.cardColors(containerColor = BgCard),
        shape = RoundedCornerShape(16.dp)
    ) {
        Column(modifier = Modifier.padding(16.dp)) {
            Row(modifier = Modifier.fillMaxWidth(), horizontalArrangement = Arrangement.SpaceBetween) {
                Text("${stringResource(R.string.subscription_status)}：${subscription.status}", style = MaterialTheme.typography.bodySmall, color = TextPrimary)
                Text("¥${subscription.price}", style = MaterialTheme.typography.bodySmall, color = TextPrimary, fontWeight = FontWeight.Bold)
            }
            Spacer(modifier = Modifier.height(4.dp))
            Text("${subscription.startTime} - ${subscription.endTime}", style = MaterialTheme.typography.bodySmall, color = TextSecondary)
        }
    }
}
