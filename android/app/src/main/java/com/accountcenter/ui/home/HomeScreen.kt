package com.accountcenter.ui.home

import androidx.compose.foundation.layout.Column
import androidx.compose.foundation.layout.Row
import androidx.compose.foundation.layout.Spacer
import androidx.compose.foundation.layout.fillMaxSize
import androidx.compose.foundation.layout.fillMaxWidth
import androidx.compose.foundation.layout.height
import androidx.compose.foundation.layout.padding
import androidx.compose.foundation.layout.size
import androidx.compose.foundation.layout.width
import androidx.compose.foundation.shape.CircleShape
import androidx.compose.material.icons.Icons
import androidx.compose.material.icons.filled.CreditCard
import androidx.compose.material.icons.filled.Info
import androidx.compose.material.icons.filled.Lock
import androidx.compose.material.icons.filled.ShoppingCart
import androidx.compose.material3.Button
import androidx.compose.material3.Card
import androidx.compose.material3.CardDefaults
import androidx.compose.material3.Icon
import androidx.compose.material3.ListItem
import androidx.compose.material3.MaterialTheme
import androidx.compose.material3.Scaffold
import androidx.compose.material3.Text
import androidx.compose.material3.TopAppBar
import androidx.compose.runtime.Composable
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.draw.clip
import androidx.compose.ui.graphics.Color
import androidx.compose.ui.graphics.vector.ImageVector
import androidx.compose.ui.text.font.FontWeight
import androidx.compose.ui.unit.dp
import androidx.hilt.navigation.compose.hiltViewModel

@Composable
fun HomeScreen(
    homeViewModel: HomeViewModel = hiltViewModel(),
    onLogout: () -> Unit
) {
    Scaffold(topBar = { TopAppBar(title = { Text("用户中心") }) }) { padding ->
        Column(
            modifier = Modifier
                .fillMaxSize()
                .padding(padding)
                .padding(16.dp)
        ) {
            Card(
                modifier = Modifier.fillMaxWidth(),
                colors = CardDefaults.cardColors(containerColor = MaterialTheme.colorScheme.primaryContainer)
            ) {
                Row(modifier = Modifier.padding(16.dp), verticalAlignment = Alignment.CenterVertically) {
                    androidx.compose.foundation.layout.Box(
                        modifier = Modifier
                            .size(64.dp)
                            .clip(CircleShape)
                            .background(MaterialTheme.colorScheme.primary)
                    ) {
                        Text("U", color = MaterialTheme.colorScheme.onPrimary, style = MaterialTheme.typography.titleLarge, fontWeight = FontWeight.Bold, modifier = Modifier.align(Alignment.Center))
                    }
                    Spacer(modifier = Modifier.width(16.dp))
                    Column {
                        Text("用户", style = MaterialTheme.typography.titleMedium, fontWeight = FontWeight.Bold)
                        Text("138****1234", style = MaterialTheme.typography.bodyMedium, color = MaterialTheme.colorScheme.onSurfaceVariant)
                    }
                }
            }
            Spacer(modifier = Modifier.height(24.dp))
            Text("功能", style = MaterialTheme.typography.titleSmall, color = MaterialTheme.colorScheme.onSurfaceVariant)
            Spacer(modifier = Modifier.height(8.dp))
            FeatureItem(Icons.Default.ShoppingCart, "订阅管理") { }
            FeatureItem(Icons.Default.CreditCard, "积分中心") { }
            FeatureItem(Icons.Default.Lock, "安全设置") { }
            FeatureItem(Icons.Default.Info, "关于") { }
            Spacer(modifier = Modifier.height(24.dp))
            Button(onClick = { homeViewModel.logout(onLogout) }, modifier = Modifier.fillMaxWidth()) {
                Text("退出登录")
            }
        }
    }
}

@Composable
fun FeatureItem(icon: ImageVector, title: String, onClick: () -> Unit) {
    ListItem(
        leadingContent = { Icon(icon, contentDescription = title, tint = MaterialTheme.colorScheme.onSurfaceVariant) },
        headlineContent = { Text(title) },
        modifier = Modifier.fillMaxWidth(),
        onClick = onClick
    )
}

private fun phoneDesensitized(phone: String): String {
    return if (phone.length == 11) "${phone.take(3)}****${phone.takeLast(4)}" else "未绑定手机号"
}

