package com.eventmpos.mpos.ui

import androidx.compose.runtime.Composable
import androidx.navigation.NavType
import androidx.navigation.compose.NavHost
import androidx.navigation.compose.composable
import androidx.navigation.compose.rememberNavController
import androidx.navigation.navArgument
import com.eventmpos.mpos.ui.cart.CartScreen
import com.eventmpos.mpos.ui.catalogue.CatalogueScreen
import com.eventmpos.mpos.ui.customer.CustomerScreen
import com.eventmpos.mpos.ui.home.HomeScreen
import com.eventmpos.mpos.ui.login.LoginScreen
import com.eventmpos.mpos.ui.orders.OrdersScreen
import com.eventmpos.mpos.ui.promotions.PromotionsScreen
import com.eventmpos.mpos.ui.receipt.ReceiptScreen

object Routes {
    const val LOGIN = "login"
    const val HOME = "home"
    const val CATALOGUE = "catalogue/{eventId}"
    const val CUSTOMER = "customer"
    const val PROMOTIONS = "promotions/{eventId}"
    const val CART = "cart/{eventId}"
    const val ORDERS = "orders/{eventId}"
    const val RECEIPT = "receipt/{orderKey}"
    fun catalogue(eventId: String) = "catalogue/$eventId"
    fun promotions(eventId: String) = "promotions/$eventId"
    fun cart(eventId: String) = "cart/$eventId"
    fun orders(eventId: String) = "orders/$eventId"
    fun receipt(orderKey: String) = "receipt/$orderKey"
}

@Composable
fun AppNavHost(startLoggedIn: Boolean) {
    val navController = rememberNavController()
    val eventArg = listOf(navArgument("eventId") { type = NavType.StringType })

    NavHost(
        navController = navController,
        startDestination = if (startLoggedIn) Routes.HOME else Routes.LOGIN,
    ) {
        composable(Routes.LOGIN) {
            LoginScreen(
                onLoggedIn = {
                    navController.navigate(Routes.HOME) { popUpTo(Routes.LOGIN) { inclusive = true } }
                },
            )
        }
        composable(Routes.HOME) {
            HomeScreen(
                onLoggedOut = {
                    navController.navigate(Routes.LOGIN) { popUpTo(Routes.HOME) { inclusive = true } }
                },
                onBrowseProducts = { navController.navigate(Routes.catalogue(it)) },
                onIdentifyCustomer = { navController.navigate(Routes.CUSTOMER) },
                onViewPromotions = { navController.navigate(Routes.promotions(it)) },
                onViewOrders = { navController.navigate(Routes.orders(it)) },
            )
        }
        composable(Routes.CATALOGUE, arguments = eventArg) { entry ->
            val eventId = entry.arguments?.getString("eventId").orEmpty()
            CatalogueScreen(onViewCart = { navController.navigate(Routes.cart(eventId)) })
        }
        composable(Routes.CUSTOMER) { CustomerScreen() }
        composable(Routes.PROMOTIONS, arguments = eventArg) { PromotionsScreen() }
        composable(Routes.CART, arguments = eventArg) { entry ->
            val eventId = entry.arguments?.getString("eventId").orEmpty()
            CartScreen(
                onConfirmed = {
                    navController.navigate(Routes.orders(eventId)) { popUpTo(Routes.HOME) }
                },
                onIdentifyCustomer = { navController.navigate(Routes.CUSTOMER) },
            )
        }
        composable(Routes.ORDERS, arguments = eventArg) {
            OrdersScreen(onOpenReceipt = { navController.navigate(Routes.receipt(it)) })
        }
        composable(
            Routes.RECEIPT,
            arguments = listOf(navArgument("orderKey") { type = NavType.StringType }),
        ) {
            ReceiptScreen()
        }
    }
}
