package com.accountcenter.ui.security

import androidx.compose.foundation.background
import androidx.compose.foundation.layout.*
import androidx.compose.foundation.lazy.LazyColumn
import androidx.compose.foundation.shape.CircleShape
import androidx.compose.foundation.shape.RoundedCornerShape
import androidx.compose.material.icons.Icons
import androidx.compose.material.icons.filled.ArrowBack
import androidx.compose.material.icons.filled.Computer
import androidx.compose.material.icons.filled.PhoneAndroid
import androidx.compose.material3.*
import androidx.compose.runtime.Composable
import androidx.compose.runtime.getValue
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.draw.clip
import androidx.compose.ui.graphics.Color
import androidx.compose.ui.text.font.FontWeight
import androidx.compose.ui.unit.dp
import androidx.hilt.navigation.compose.hiltViewModel
import androidx.lifecycle.compose.collectAsStateWithLifecycle
import com.accountcenter.model.PushDevice
import com.accountcenter.model.RiskEvent
import com.accountcenter.ui.theme.*

@OptIn(ExperimentalMaterial3Api::class)
@Composable
fun SecurityScreen(
    viewModel: SecurityViewModel = hiltViewModel(),
    onNavigateBack: () -> Unit
) {
    val riskEvents by viewModel.riskEvents.collectAsStateWithLifecycle()
    val devices by viewModel.devices.collectAsStateWithLifecycle()
    val isLoading by viewModel.isLoading.collectAsStateWithLifecycle()

    Box(modifier = Modifier.fillMaxSize().background(BgPrimary)) {
        Column(modifier = Modifier.fillMaxSize()) {
            CenterAlignedTopAppBar(
                title = { Text("安全设置", color = TextPrimary) },
                navigationIcon = {
                    IconButton(onClick = onNavigateBack) {
                        Icon(Icons.Default.ArrowBack, contentDescription = "返回", tint = TextPrimary)
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
                        Text("风险事件", style = MaterialTheme.typography.labelSmall, color = TextSecondary)
                    }
                    riskEvents.forEach { event ->
                        item {
                            RiskEventCard(event)
                        }
                    }

                    item {
                        Spacer(modifier = Modifier.height(8.dp))
                        Text("登录设备", style = MaterialTheme.typography.labelSmall, color = TextSecondary)
                    }
                    devices.forEach { device ->
                        item {
                            DeviceCard(device)
                        }
                    }
                }
            }
        }
    }
}

@Composable
private fun RiskEventCard(event: RiskEvent) {
    val levelColor = when (event.riskLevel) {
        "critical" -> Danger
        "high" -> Color(0xFFFF9800)
        "medium" -> Color(0xFFFFEB3B)
        "low" -> Success
        else -> TextSecondary
    }
    Card(
        modifier = Modifier.fillMaxWidth(),
        colors = CardDefaults.cardColors(containerColor = BgCard),
        shape = RoundedCornerShape(16.dp)
    ) {
        Row(
            modifier = Modifier.padding(16.dp),
            verticalAlignment = Alignment.CenterVertically
        ) {
            Box(
                modifier = Modifier.size(12.dp).clip(CircleShape).background(levelColor)
            )
            Spacer(modifier = Modifier.width(12.dp))
            Column(modifier = Modifier.weight(1f)) {
                Text(event.eventType, style = MaterialTheme.typography.bodyMedium, color = TextPrimary, fontWeight = FontWeight.Bold)
                Text(event.createdAt, style = MaterialTheme.typography.bodySmall, color = TextSecondary)
            }
            Spacer(modifier = Modifier.width(8.dp))
            Surface(
                color = levelColor.copy(alpha = 0.15f),
                shape = RoundedCornerShape(6.dp)
            ) {
                Text(
                    event.riskLevel,
                    modifier = Modifier.padding(horizontal = 6.dp, vertical = 2.dp),
                    style = MaterialTheme.typography.labelSmall,
                    color = levelColor
                )
            }
        }
    }
}

@Composable
private fun DeviceCard(device: PushDevice) {
    Card(
        modifier = Modifier.fillMaxWidth(),
        colors = CardDefaults.cardColors(containerColor = BgCard),
        shape = RoundedCornerShape(16.dp)
    ) {
        Row(
            modifier = Modifier.padding(16.dp).fillMaxWidth(),
            verticalAlignment = Alignment.CenterVertically
        ) {
            Icon(
                if (device.platform == "web") Icons.Default.Computer else Icons.Default.PhoneAndroid,
                contentDescription = null,
                tint = BrandPrimary,
                modifier = Modifier.size(24.dp)
            )
            Spacer(modifier = Modifier.width(12.dp))
            Column(modifier = Modifier.weight(1f)) {
                Text(device.deviceName ?: device.platform, style = MaterialTheme.typography.bodyMedium, color = TextPrimary)
                Text("最近活跃", style = MaterialTheme.typography.bodySmall, color = TextSecondary)
            }
            Text(
                if (device.isActive) "活跃中" else "离线",
                style = MaterialTheme.typography.bodySmall,
                color = if (device.isActive) Success else TextSecondary,
                fontWeight = FontWeight.Medium
            )
        }
    }
}
