package com.accountcenter.ui.register

import androidx.compose.foundation.background
import androidx.compose.foundation.layout.*
import androidx.compose.foundation.rememberScrollState
import androidx.compose.foundation.text.KeyboardOptions
import androidx.compose.foundation.verticalScroll
import androidx.compose.material.icons.Icons
import androidx.compose.material.icons.filled.Person
import androidx.compose.material3.*
import androidx.compose.runtime.*
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.res.stringResource
import androidx.compose.ui.text.input.KeyboardType
import androidx.compose.ui.text.input.PasswordVisualTransformation
import androidx.compose.ui.unit.dp
import androidx.hilt.navigation.compose.hiltViewModel
import com.accountcenter.R
import com.accountcenter.ui.components.GradientButton
import com.accountcenter.ui.theme.*

@Composable
fun RegisterScreen(
    viewModel: RegisterViewModel = hiltViewModel(),
    onNavigateBack: () -> Unit,
    onRegisterSuccess: () -> Unit
) {
    val uiState = viewModel.uiState.collectAsState().value
    Box(modifier = Modifier.fillMaxSize().background(BgPrimary)) {
        Column(
            modifier = Modifier
                .fillMaxSize()
                .verticalScroll(rememberScrollState())
                .padding(32.dp),
            horizontalAlignment = Alignment.CenterHorizontally
        ) {
            Spacer(modifier = Modifier.height(24.dp))
            Column(modifier = Modifier.padding(bottom = 32.dp), horizontalAlignment = Alignment.CenterHorizontally) {
                Icon(
                    imageVector = Icons.Default.Person,
                    contentDescription = null,
                    modifier = Modifier.size(56.dp),
                    tint = BrandPrimary
                )
                Spacer(modifier = Modifier.height(12.dp))
                Text(stringResource(R.string.register_create_account), style = MaterialTheme.typography.headlineLarge, color = TextPrimary)
                Text(stringResource(R.string.register_subtitle), style = MaterialTheme.typography.bodyMedium, color = TextSecondary)
            }

            OutlinedTextField(
                value = uiState.phoneNumber,
                onValueChange = viewModel::onPhoneNumberChange,
                label = { Text(stringResource(R.string.register_phone_label)) },
                singleLine = true,
                keyboardOptions = KeyboardOptions(keyboardType = KeyboardType.Phone),
                modifier = Modifier.fillMaxWidth(),
                colors = OutlinedTextFieldDefaults.colors(focusedBorderColor = BrandSecondary, unfocusedBorderColor = Divider, focusedLabelColor = BrandSecondary, unfocusedLabelColor = TextSecondary, cursorColor = BrandSecondary)
            )
            Spacer(modifier = Modifier.height(16.dp))

            Row(modifier = Modifier.fillMaxWidth(), horizontalArrangement = Arrangement.spacedBy(8.dp)) {
                OutlinedTextField(
                    value = uiState.verificationCode,
                    onValueChange = viewModel::onVerificationCodeChange,
                    label = { Text(stringResource(R.string.register_code_label)) },
                    singleLine = true,
                    keyboardOptions = KeyboardOptions(keyboardType = KeyboardType.Number),
                    modifier = Modifier.weight(1f),
                    colors = OutlinedTextFieldDefaults.colors(focusedBorderColor = BrandSecondary, unfocusedBorderColor = Divider, focusedLabelColor = BrandSecondary, unfocusedLabelColor = TextSecondary, cursorColor = BrandSecondary)
                )
                val btnText = if (uiState.countdownSeconds > 0) "${uiState.countdownSeconds}s" else stringResource(R.string.register_send_code)
                OutlinedButton(
                    onClick = viewModel::sendVerificationCode,
                    enabled = !uiState.isLoading && uiState.countdownSeconds == 0 && uiState.phoneNumber.isNotEmpty()
                ) { Text(btnText, color = BrandSecondary) }
            }
            Spacer(modifier = Modifier.height(16.dp))

            OutlinedTextField(
                value = uiState.accountId,
                onValueChange = viewModel::onAccountIdChange,
                label = { Text(stringResource(R.string.register_account_id)) },
                singleLine = true,
                modifier = Modifier.fillMaxWidth(),
                colors = OutlinedTextFieldDefaults.colors(focusedBorderColor = BrandSecondary, unfocusedBorderColor = Divider, focusedLabelColor = BrandSecondary, unfocusedLabelColor = TextSecondary, cursorColor = BrandSecondary)
            )
            Spacer(modifier = Modifier.height(16.dp))

            OutlinedTextField(
                value = uiState.password,
                onValueChange = viewModel::onPasswordChange,
                label = { Text(stringResource(R.string.register_password_label)) },
                singleLine = true,
                visualTransformation = PasswordVisualTransformation(),
                modifier = Modifier.fillMaxWidth(),
                colors = OutlinedTextFieldDefaults.colors(focusedBorderColor = BrandSecondary, unfocusedBorderColor = Divider, focusedLabelColor = BrandSecondary, unfocusedLabelColor = TextSecondary, cursorColor = BrandSecondary)
            )
            Spacer(modifier = Modifier.height(16.dp))

            OutlinedTextField(
                value = uiState.confirmPassword,
                onValueChange = viewModel::onConfirmPasswordChange,
                label = { Text(stringResource(R.string.register_confirm_password)) },
                singleLine = true,
                visualTransformation = PasswordVisualTransformation(),
                modifier = Modifier.fillMaxWidth(),
                colors = OutlinedTextFieldDefaults.colors(focusedBorderColor = BrandSecondary, unfocusedBorderColor = Divider, focusedLabelColor = BrandSecondary, unfocusedLabelColor = TextSecondary, cursorColor = BrandSecondary)
            )
            Spacer(modifier = Modifier.height(16.dp))

            OutlinedTextField(
                value = uiState.referralCode,
                onValueChange = viewModel::onReferralCodeChange,
                label = { Text(stringResource(R.string.register_referral_code)) },
                singleLine = true,
                modifier = Modifier.fillMaxWidth(),
                colors = OutlinedTextFieldDefaults.colors(focusedBorderColor = BrandSecondary, unfocusedBorderColor = Divider, focusedLabelColor = BrandSecondary, unfocusedLabelColor = TextSecondary, cursorColor = BrandSecondary)
            )
            Spacer(modifier = Modifier.height(8.dp))

            Row(verticalAlignment = Alignment.CenterVertically) {
                Checkbox(
                    checked = uiState.agreeToTerms,
                    onCheckedChange = viewModel::onAgreeToTermsChange,
                    colors = CheckboxDefaults.colors(checkedColor = BrandPrimary)
                )
                Text(stringResource(R.string.register_agree_terms), style = MaterialTheme.typography.bodySmall, color = TextSecondary)
            }

            uiState.errorMessage?.let {
                Text(it, color = Danger, style = MaterialTheme.typography.bodySmall)
            }
            Spacer(modifier = Modifier.height(16.dp))

            GradientButton(
                text = stringResource(R.string.register_button),
                onClick = { viewModel.register(onRegisterSuccess) },
                enabled = !uiState.isLoading
            )

            Spacer(modifier = Modifier.height(16.dp))
            Row(modifier = Modifier.fillMaxWidth(), horizontalArrangement = Arrangement.Center) {
                Text(stringResource(R.string.register_has_account), style = MaterialTheme.typography.bodySmall, color = TextSecondary)
                TextButton(onClick = onNavigateBack) {
                    Text(stringResource(R.string.register_login_now), color = BrandSecondary)
                }
            }
            Spacer(modifier = Modifier.height(24.dp))
        }

        if (uiState.isLoading) {
            CircularProgressIndicator(modifier = Modifier.align(Alignment.Center), color = BrandSecondary)
        }
    }
}
