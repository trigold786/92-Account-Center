package com.accountcenter.ui.home

import androidx.compose.foundation.background
import androidx.compose.foundation.layout.*
import androidx.compose.foundation.shape.CircleShape
import androidx.compose.foundation.shape.RoundedCornerShape
import androidx.compose.material.icons.Icons
import androidx.compose.material.icons.filled.*
import androidx.compose.material3.*
import androidx.compose.runtime.*
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.draw.clip
import androidx.compose.ui.graphics.Brush
import androidx.compose.ui.graphics.Color
import androidx.compose.ui.graphics.vector.ImageVector
import androidx.compose.ui.res.stringResource
import androidx.compose.ui.text.font.FontWeight
import androidx.compose.ui.unit.dp
import androidx.hilt.navigation.compose.hiltViewModel
import androidx.lifecycle.compose.collectAsStateWithLifecycle
import com.accountcenter.R
import com.accountcenter.ui.components.AppCard
import com.accountcenter.ui.theme.*

@OptIn(ExperimentalMaterial3Api::class)
@Composable
fun HomeScreen(
    homeViewModel: HomeViewModel = hiltViewModel(),
    onLogout: () -> Unit,
    onNavigateToSubscription: () -> Unit = {},
    onNavigateToCredits: () -> Unit = {},
    onNavigateToSecurity: () -> Unit = {},
    onNavigateToAbout: () -> Unit = {}
) {
    val userDisplay by homeViewModel.userDisplay.collectAsStateWithLifecycle()
    val rfmScore by homeViewModel.rfmScore.collectAsStateWithLifecycle()

    Box(modifier = Modifier.fillMaxSize().background(BgPrimary)) {
        Column(
            modifier = Modifier.fillMaxSize(),
            verticalArrangement = Arrangement.Top
        ) {
            CenterAlignedTopAppBar(
                title = { Text(stringResource(R.string.home_title), color = TextPrimary) },
                colors = TopAppBarDefaults.centerAlignedTopAppBarColors(
                    containerColor = BgPrimary.copy(alpha = 0.9f)
                )
            )

            Column(
                modifier = Modifier.fillMaxSize().padding(16.dp),
                verticalArrangement = Arrangement.spacedBy(12.dp)
            ) {
                // User card
                AppCard(
                    modifier = Modifier.fillMaxWidth()
                ) {
                    Row(
                        modifier = Modifier.fillMaxWidth(),
                        verticalAlignment = Alignment.CenterVertically
                    ) {
                        Box(
                            modifier = Modifier
                                .size(64.dp)
                                .clip(CircleShape)
                                .background(
                                    Brush.horizontalGradient(listOf(BrandPrimary, BrandSecondary))
                                ),
                            contentAlignment = Alignment.Center
                        ) {
                            Text(
                                userDisplay.accountId.firstOrNull()?.uppercase() ?: "U",
                                style = MaterialTheme.typography.titleLarge,
                                fontWeight = FontWeight.Bold,
                                color = Color.White
                            )
                        }
                        Spacer(modifier = Modifier.width(16.dp))
                        Column {
                            Text(
                                userDisplay.accountId.ifEmpty { stringResource(R.string.user_placeholder) },
                                style = MaterialTheme.typography.titleMedium,
                                color = TextPrimary
                            )
                            Text(
                                userDisplay.phoneNumber.ifEmpty { stringResource(R.string.phone_unbound) },
                                style = MaterialTheme.typography.bodySmall,
                                color = TextSecondary
                            )
                            Row(verticalAlignment = Alignment.CenterVertically) {
                                Box(
                                    modifier = Modifier.size(6.dp).clip(CircleShape).background(BrandSecondary)
                                )
                                Spacer(modifier = Modifier.width(4.dp))
                                Text("Lv.2", style = MaterialTheme.typography.labelSmall, color = BrandSecondary)
                            }
                        }
                    }
                }

                // RFM card
                rfmScore?.let { score ->
                    AppCard(
                        modifier = Modifier.fillMaxWidth()
                    ) {
                        Row(verticalAlignment = Alignment.CenterVertically) {
                            Text("\uD83C\uDFAF", style = MaterialTheme.typography.titleMedium)
                            Spacer(modifier = Modifier.width(12.dp))
                            Column {
                                Text(score.rfmSegmentCn, style = MaterialTheme.typography.titleSmall, fontWeight = FontWeight.Bold, color = TextPrimary)
                                Text("RFM ${score.rfmSegment}", style = MaterialTheme.typography.bodySmall, color = TextSecondary)
                            }
                        }
                    }
                }

                // Feature list
                Text(stringResource(R.string.features_section), style = MaterialTheme.typography.labelSmall, color = TextSecondary)
                Card(
                    modifier = Modifier.fillMaxWidth(),
                    colors = CardDefaults.cardColors(containerColor = BgCard),
                    shape = RoundedCornerShape(16.dp)
                ) {
                    Column {
                        FeatureRow(Icons.Default.ShoppingCart, stringResource(R.string.feature_subscription), onNavigateToSubscription)
                        HorizontalDivider(color = Divider)
                        FeatureRow(Icons.Default.AccountBalanceWallet, stringResource(R.string.feature_credits), onNavigateToCredits)
                        HorizontalDivider(color = Divider)
                        FeatureRow(Icons.Default.Lock, stringResource(R.string.feature_security), onNavigateToSecurity)
                        HorizontalDivider(color = Divider)
                        FeatureRow(Icons.Default.Info, stringResource(R.string.feature_about), onNavigateToAbout)
                    }
                }

                // Logout
                Button(
                    onClick = { homeViewModel.logout(onLogout) },
                    modifier = Modifier.fillMaxWidth().height(52.dp),
                    colors = ButtonDefaults.buttonColors(containerColor = BgCard),
                    shape = RoundedCornerShape(12.dp)
                ) {
                    Text(stringResource(R.string.logout), color = Danger)
                }
            }
        }
    }
}

@Composable
private fun FeatureRow(icon: ImageVector, label: String, onClick: () -> Unit) {
    Surface(
        onClick = onClick,
        color = Color.Transparent,
        modifier = Modifier.fillMaxWidth().height(56.dp)
    ) {
        Row(
            modifier = Modifier.fillMaxSize().padding(horizontal = 16.dp),
            verticalAlignment = Alignment.CenterVertically
        ) {
            Icon(icon, contentDescription = label, tint = BrandPrimary, modifier = Modifier.size(24.dp))
            Spacer(modifier = Modifier.width(12.dp))
            Text(label, style = MaterialTheme.typography.bodyLarge, color = TextPrimary, modifier = Modifier.weight(1f))
            Icon(Icons.Default.KeyboardArrowRight, contentDescription = null, tint = TextSecondary, modifier = Modifier.size(16.dp))
        }
    }
}
