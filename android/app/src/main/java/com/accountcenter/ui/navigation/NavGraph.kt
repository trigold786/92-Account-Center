package com.accountcenter.ui.navigation

import androidx.compose.runtime.Composable
import androidx.compose.runtime.collectAsState
import androidx.compose.runtime.getValue
import androidx.navigation.NavHostController
import androidx.navigation.compose.NavHost
import androidx.navigation.compose.composable
import androidx.navigation.compose.rememberNavController
import com.accountcenter.repository.AuthRepository
import com.accountcenter.ui.about.AboutScreen
import com.accountcenter.ui.credits.CreditsScreen
import com.accountcenter.ui.home.HomeScreen
import com.accountcenter.ui.login.LoginScreen
import com.accountcenter.ui.register.RegisterScreen
import com.accountcenter.ui.security.SecurityScreen
import com.accountcenter.ui.subscription.SubscriptionScreen

sealed class Screen(val route: String) {
    object Login : Screen("login")
    object Register : Screen("register")
    object Home : Screen("home")
    object Subscription : Screen("subscription")
    object Credits : Screen("credits")
    object Security : Screen("security")
    object About : Screen("about")
}

@Composable
fun NavGraph(
    navController: NavHostController = rememberNavController(),
    authRepository: AuthRepository
) {
    val isAuthenticated by authRepository.isAuthenticated.collectAsState(initial = false)

    val startDestination = if (isAuthenticated) Screen.Home.route else Screen.Login.route

    NavHost(navController = navController, startDestination = startDestination) {
        composable(Screen.Login.route) {
            LoginScreen(
                onNavigateToRegister = { navController.navigate(Screen.Register.route) },
                onLoginSuccess = {
                    navController.navigate(Screen.Home.route) {
                        popUpTo(0)
                    }
                }
            )
        }
        composable(Screen.Register.route) {
            RegisterScreen(
                onNavigateBack = { navController.popBackStack() },
                onRegisterSuccess = {
                    navController.navigate(Screen.Home.route) {
                        popUpTo(0)
                    }
                }
            )
        }
        composable(Screen.Home.route) {
            HomeScreen(
                onLogout = {
                    navController.navigate(Screen.Login.route) {
                        popUpTo(0)
                    }
                },
                onNavigateToSubscription = { navController.navigate(Screen.Subscription.route) },
                onNavigateToCredits = { navController.navigate(Screen.Credits.route) },
                onNavigateToSecurity = { navController.navigate(Screen.Security.route) },
                onNavigateToAbout = { navController.navigate(Screen.About.route) }
            )
        }
        composable(Screen.Subscription.route) {
            SubscriptionScreen(
                onNavigateBack = { navController.popBackStack() }
            )
        }
        composable(Screen.Credits.route) {
            CreditsScreen(
                onNavigateBack = { navController.popBackStack() }
            )
        }
        composable(Screen.Security.route) {
            SecurityScreen(
                onNavigateBack = { navController.popBackStack() }
            )
        }
        composable(Screen.About.route) {
            AboutScreen(
                onNavigateBack = { navController.popBackStack() }
            )
        }
    }
}
