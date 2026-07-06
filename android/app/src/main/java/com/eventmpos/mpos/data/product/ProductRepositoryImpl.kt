package com.eventmpos.mpos.data.product

import com.eventmpos.mpos.data.auth.AuthApi
import com.eventmpos.mpos.data.auth.VariantDto
import com.eventmpos.mpos.data.local.ProductDao
import com.eventmpos.mpos.data.local.ProductEntity
import com.eventmpos.mpos.domain.product.Product
import com.eventmpos.mpos.domain.product.ProductRepository
import com.eventmpos.mpos.domain.product.ProductVariant
import kotlinx.coroutines.Dispatchers
import kotlinx.coroutines.flow.Flow
import kotlinx.coroutines.flow.map
import kotlinx.coroutines.withContext
import kotlinx.serialization.builtins.ListSerializer
import kotlinx.serialization.json.Json
import javax.inject.Inject
import javax.inject.Singleton

@Singleton
class ProductRepositoryImpl @Inject constructor(
    private val dao: ProductDao,
    private val api: AuthApi,
    private val json: Json,
) : ProductRepository {

    override fun observeProducts(eventId: String): Flow<List<Product>> =
        dao.observeForEvent(eventId).map { list -> list.map { it.toDomain() } }

    override suspend fun findByBarcode(eventId: String, barcode: String): Product? =
        dao.findByBarcode(eventId, barcode.trim())?.toDomain()

    override suspend fun getProduct(eventId: String, productId: String): Product? =
        dao.find(eventId, productId)?.toDomain()

    override suspend fun refreshStock(eventId: String) {
        withContext(Dispatchers.IO) {
            runCatching {
                api.getStockLevels(eventId).forEach { level ->
                    dao.updateRemaining(eventId, level.productId, level.remaining)
                }
            } // offline / server error -> keep last cached levels
        }
    }

    private fun ProductEntity.toDomain() = Product(
        id = id,
        sku = sku,
        name = name,
        brand = brand,
        price = price,
        imageUrl = imageUrl,
        barcodes = barcodes.split("\n").filter { it.isNotBlank() },
        remainingStock = remainingStock,
        variants = runCatching {
            json.decodeFromString(ListSerializer(VariantDto.serializer()), variantsJson)
                .map { ProductVariant(it.id, it.name, it.price) }
        }.getOrDefault(emptyList()),
    )
}
