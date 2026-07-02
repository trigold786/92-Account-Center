package com.accountcenter.ui.about

import android.content.pm.PackageManager
import androidx.compose.foundation.background
import androidx.compose.foundation.layout.*
import androidx.compose.foundation.shape.RoundedCornerShape
import androidx.compose.material.icons.Icons
import androidx.compose.material.icons.filled.ArrowBack
import androidx.compose.material.icons.filled.KeyboardArrowRight
import androidx.compose.material.icons.filled.Person
import androidx.compose.material3.*
import androidx.compose.runtime.Composable
import androidx.compose.runtime.remember
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.graphics.Color
import androidx.compose.ui.platform.LocalContext
import androidx.compose.ui.res.stringResource
import androidx.compose.ui.text.font.FontWeight
import androidx.compose.ui.unit.dp
import com.accountcenter.R
import com.accountcenter.ui.theme.*

@OptIn(ExperimentalMaterial3Api::class)
@Composable
fun AboutScreen(onNavigateBack: () -> Unit) {
    val context = LocalContext.current
    val packageInfo = remember {
        try {
            context.packageManager.getPackageInfo(context.packageName, 0)
        } catch (e: PackageManager.NameNotFoundException) {
            null
        }
    }
    val versionName = packageInfo?.versionName ?: "未知"
    val versionCode = if (android.os.Build.VERSION.SDK_INT >= 28) {
        packageInfo?.longVersionCode ?: 0
    } else {
        @Suppress("DEPRECATION")
        packageInfo?.versionCode ?: 0
    }

    Box(modifier = Modifier.fillMaxSize().background(BgPrimary)) {
        Column(modifier = Modifier.fillMaxSize()) {
            CenterAlignedTopAppBar(
                title = { Text(stringResource(R.string.about_title), color = TextPrimary) },
                navigationIcon = {
                    IconButton(onClick = onNavigateBack) {
                        Icon(Icons.Default.ArrowBack, contentDescription = stringResource(R.string.back), tint = TextPrimary)
                    }
                },
                colors = TopAppBarDefaults.centerAlignedTopAppBarColors(
                    containerColor = BgPrimary.copy(alpha = 0.9f)
                )
            )

            Column(
                modifier = Modifier.fillMaxSize().padding(32.dp),
                horizontalAlignment = Alignment.CenterHorizontally
            ) {
                Spacer(modifier = Modifier.height(24.dp))

                Icon(
                    imageVector = Icons.Default.Person,
                    contentDescription = null,
                    modifier = Modifier.size(56.dp),
                    tint = BrandPrimary
                )
                Spacer(modifier = Modifier.height(12.dp))
                Text(stringResource(R.string.about_account_center), style = MaterialTheme.typography.headlineLarge, color = TextPrimary)
                Text("Version $versionName Build $versionCode", style = MaterialTheme.typography.bodySmall, color = TextSecondary)
                Spacer(modifier = Modifier.height(32.dp))

                Card(
                    modifier = Modifier.fillMaxWidth(),
                    colors = CardDefaults.cardColors(containerColor = BgCard),
                    shape = RoundedCornerShape(16.dp)
                ) {
                    Column {
                        InfoRow(stringResource(R.string.about_app_version), versionName)
                        HorizontalDivider(color = Divider)
                        InfoRow(stringResource(R.string.about_build), versionCode.toString())
                    }
                }
                Spacer(modifier = Modifier.height(12.dp))

                Card(
                    modifier = Modifier.fillMaxWidth(),
                    colors = CardDefaults.cardColors(containerColor = BgCard),
                    shape = RoundedCornerShape(16.dp)
                ) {
                    Column {
                        LegalRow(stringResource(R.string.about_terms))
                        HorizontalDivider(color = Divider)
                        LegalRow(stringResource(R.string.about_privacy))
                    }
                }

                Spacer(modifier = Modifier.height(32.dp))
                Text("© 2024 Account Center", style = MaterialTheme.typography.labelSmall, color = TextSecondary)
            }
        }
    }
}

@Composable
private fun InfoRow(label: String, value: String) {
    Row(
        modifier = Modifier.fillMaxWidth().padding(horizontal = 16.dp, vertical = 14.dp),
        horizontalArrangement = Arrangement.SpaceBetween,
        verticalAlignment = Alignment.CenterVertically
    ) {
        Text(label, style = MaterialTheme.typography.bodyMedium, color = TextPrimary)
        Text(value, style = MaterialTheme.typography.bodyMedium, color = TextSecondary)
    }
}

@Composable
private fun LegalRow(label: String) {
    Surface(
        onClick = {},
        color = Color.Transparent,
        modifier = Modifier.fillMaxWidth()
    ) {
        Row(
            modifier = Modifier.fillMaxWidth().padding(horizontal = 16.dp, vertical = 14.dp),
            horizontalArrangement = Arrangement.SpaceBetween,
            verticalAlignment = Alignment.CenterVertically
        ) {
            Text(label, style = MaterialTheme.typography.bodyMedium, color = BrandSecondary)
            Icon(
                Icons.Default.KeyboardArrowRight,
                contentDescription = null,
                tint = TextSecondary,
                modifier = Modifier.size(16.dp)
            )
        }
    }
}
