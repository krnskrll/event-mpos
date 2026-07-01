package com.eventmpos.mpos.ui

import androidx.compose.foundation.layout.padding
import androidx.compose.material3.Scaffold
import androidx.compose.runtime.Composable
import androidx.compose.runtime.getValue
import androidx.hilt.navigation.compose.hiltViewModel
import androidx.lifecycle.compose.collectAsStateWithLifecycle
import androidx.navigation.NavType
import androidx.navigation.compose.NavHost
import androidx.navigation.compose.composable
import androidx.navigation.compose.currentBackStackEntryAsState
import androidx.navigation.compose.rememberNavController
import androidx.navigation.navArgument
import com.eventmpos.mpos.ui.cart.CartScreen
import com.eventmpos.mpos.ui.catalogue.BarcodeScanScreen
import com.eventmpos.mpos.ui.catalogue.CatalogueScreen
import com.eventmpos.mpos.ui.catalogue.ProductDetailScreen
import com.eventmpos.mpos.ui.customer.CustomerScreen
import com.eventmpos.mpos.ui.home.HomeScreen
import com.eventmpos.mpos.ui.login.LoginScreen
import com.eventmpos.mpos.ui.orders.OrdersScreen
import com.eventmpos.mpos.ui.receipt.ReceiptScreen
import com.eventmpos.mpos.ui.session.EventSelectionScreen
import com.eventmpos.mpos.ui.session.SyncScreen

object Routes {
    const val LOGIN = "login"
    const val EVENT_SELECTION = "event-selection"
    const val SYNC = "sync"
    const val HOME = "home"
    const val CATALOGUE = "catalogue"
    const val CART = "cart"
    const val ORDERS = "orders"
    const val PRODUCT_DETAIL = "product-detail/{productId}"
    const val CUSTOMER = "customer"
    const val SCAN_BARCODE = "scan-barcode"
    const val RECEIPT = "receipt/{orderKey}"
    fun productDetail(id: String) = "product-detail/$id"
    fun receipt(orderKey: String) = "receipt/$orderKey"
    val TABS = setOf(HOME, CATALOGUE, CART, ORDERS)
}

@Composable
fun AppNavHost(startLoggedIn: Boolean) {
    val nav = rememberNavController()
    val backStack by nav.currentBackStackEntryAsState()
    val route = backStack?.destination?.route
    val cartBadge: CartBadgeViewModel = hiltViewModel()
    val cartCount by cartBadge.count.collectAsStateWithLifecycle()
    val connectivity: com.eventmpos.mpos.ui.components.ConnectivityViewModel = hiltViewModel()
    val isOnline by connectivity.isOnline.collectAsStateWithLifecycle()

    // A returning BA goes to event selection (re-pick + re-sync); fresh launch → login.
    val start = if (startLoggedIn) Routes.EVENT_SELECTION else Routes.LOGIN

    fun toTab(target: String) = nav.navigate(target) {
        popUpTo(Routes.HOME) { saveState = true }
        launchSingleTop = true
        restoreState = true
    }

    Scaffold(
        bottomBar = { if (route in Routes.TABS) BottomBar(route, cartCount, ::toTab) },
    ) { padding ->
      androidx.compose.foundation.layout.Column(androidx.compose.ui.Modifier.padding(padding)) {
        com.eventmpos.mpos.ui.components.OfflineBanner(isOnline)
        NavHost(navController = nav, startDestination = start, modifier = androidx.compose.ui.Modifier.weight(1f)) {
            composable(Routes.LOGIN) {
                LoginScreen(onLoggedIn = { nav.navigate(Routes.EVENT_SELECTION) { popUpTo(Routes.LOGIN) { inclusive = true } } })
            }
            composable(Routes.EVENT_SELECTION) {
                EventSelectionScreen(
                    onStarted = { nav.navigate(Routes.SYNC) },
                    onLoggedOut = { nav.navigate(Routes.LOGIN) { popUpTo(0) } },
                )
            }
            composable(Routes.SYNC) {
                SyncScreen(onSynced = { nav.navigate(Routes.HOME) { popUpTo(Routes.EVENT_SELECTION) { inclusive = true } } })
            }
            composable(Routes.HOME) {
                HomeScreen(
                    onCatalogue = { toTab(Routes.CATALOGUE) },
                    onOrders = { toTab(Routes.ORDERS) },
                    onLoggedOut = { nav.navigate(Routes.LOGIN) { popUpTo(0) } },
                )
            }
            composable(Routes.CATALOGUE) {
                CatalogueScreen(
                    onProductSelect = { nav.navigate(Routes.productDetail(it)) },
                    onCustomerScan = { nav.navigate(Routes.CUSTOMER) },
                    onScanBarcode = { nav.navigate(Routes.SCAN_BARCODE) },
                )
            }
            composable(Routes.CART) {
                CartScreen(
                    onConfirmed = { nav.navigate(Routes.ORDERS) { popUpTo(Routes.HOME) } },
                    onIdentifyCustomer = { nav.navigate(Routes.CUSTOMER) },
                )
            }
            composable(Routes.ORDERS) {
                OrdersScreen(onOpenReceipt = { nav.navigate(Routes.receipt(it)) })
            }
            composable(Routes.PRODUCT_DETAIL, arguments = listOf(navArgument("productId") { type = NavType.StringType })) {
                ProductDetailScreen(onBack = { nav.popBackStack() }, onViewCart = { toTab(Routes.CART) })
            }
            composable(Routes.CUSTOMER) {
                CustomerScreen(onBack = { nav.popBackStack() }, onDone = { nav.popBackStack() })
            }
            composable(Routes.SCAN_BARCODE) {
                BarcodeScanScreen(onClose = { nav.popBackStack() }, onViewCart = { toTab(Routes.CART) })
            }
            composable(Routes.RECEIPT, arguments = listOf(navArgument("orderKey") { type = NavType.StringType })) {
                ReceiptScreen(
                    onNewOrder = { nav.navigate(Routes.CATALOGUE) { popUpTo(Routes.HOME) } },
                    onViewOrders = { nav.navigate(Routes.ORDERS) { popUpTo(Routes.HOME) } },
                )
            }
        }
      }
    }
}
