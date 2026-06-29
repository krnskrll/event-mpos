package com.eventmpos.mpos.ui

import androidx.compose.runtime.Composable
import androidx.navigation.NavType
import androidx.navigation.compose.NavHost
import androidx.navigation.compose.composable
import androidx.navigation.compose.rememberNavController
import androidx.navigation.navArgument
import com.eventmpos.mpos.ui.catalogue.CatalogueScreen
import com.eventmpos.mpos.ui.customer.CustomerScreen
import com.eventmpos.mpos.ui.home.HomeScreen
import com.eventmpos.mpos.ui.login.LoginScreen

object Routes {
    const val LOGIN = "login"
    const val HOME = "home"
    const val CATALOGUE = "catalogue/{eventId}"
    const val CUSTOMER = "customer"
    fun catalogue(eventId: String) = "catalogue/$eventId"
}

@Composable
fun AppNavHost(startLoggedIn: Boolean) {
    val navController = rememberNavController()

    NavHost(
        navController = navController,
        startDestination = if (startLoggedIn) Routes.HOME else Routes.LOGIN,
    ) {
        composable(Routes.LOGIN) {
            LoginScreen(
                onLoggedIn = {
                    navController.navigate(Routes.HOME) {
                        popUpTo(Routes.LOGIN) { inclusive = true }
                    }
                },
            )
        }
        composable(Routes.HOME) {
            HomeScreen(
                onLoggedOut = {
                    navController.navigate(Routes.LOGIN) {
                        popUpTo(Routes.HOME) { inclusive = true }
                    }
                },
                onBrowseProducts = { eventId -> navController.navigate(Routes.catalogue(eventId)) },
                onIdentifyCustomer = { navController.navigate(Routes.CUSTOMER) },
            )
        }
        composable(
            Routes.CATALOGUE,
            arguments = listOf(navArgument("eventId") { type = NavType.StringType }),
        ) {
            CatalogueScreen()
        }
        composable(Routes.CUSTOMER) {
            CustomerScreen()
        }
    }
}
