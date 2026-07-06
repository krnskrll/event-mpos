package com.eventmpos.mpos.core.scanner

import android.Manifest
import android.content.pm.PackageManager
import android.util.Size
import androidx.activity.compose.rememberLauncherForActivityResult
import androidx.activity.result.contract.ActivityResultContracts
import androidx.annotation.OptIn
import androidx.camera.core.CameraSelector
import androidx.camera.core.ExperimentalGetImage
import androidx.camera.core.ImageAnalysis
import androidx.camera.core.Preview
import androidx.camera.lifecycle.ProcessCameraProvider
import androidx.camera.view.PreviewView
import androidx.compose.runtime.Composable
import androidx.compose.runtime.DisposableEffect
import androidx.compose.runtime.getValue
import androidx.compose.runtime.mutableStateOf
import androidx.compose.runtime.remember
import androidx.compose.runtime.rememberUpdatedState
import androidx.compose.runtime.setValue
import androidx.compose.ui.Modifier
import androidx.compose.ui.platform.LocalContext
import androidx.compose.ui.platform.LocalLifecycleOwner
import java.util.concurrent.Executors
import androidx.compose.ui.viewinterop.AndroidView
import androidx.core.content.ContextCompat
import com.google.mlkit.vision.barcode.BarcodeScannerOptions
import com.google.mlkit.vision.barcode.BarcodeScanning
import com.google.mlkit.vision.barcode.common.Barcode
import com.google.mlkit.vision.common.InputImage

/**
 * Reusable camera scanner: a CameraX preview with an ML Kit barcode/QR analyzer bound to the
 * composition's lifecycle. Decoded values are delivered to [onDecoded]. This is the *camera source*
 * for the [ScannerPort] pipeline — on a Sunmi device the hardware laser scanner replaces it, but the
 * downstream selling flow is identical.
 *
 * Works on the Android emulator: set the AVD's back camera to "Webcam0" (or "VirtualScene" and point
 * it at a printed code) and ML Kit decodes live frames the same way it would on a physical phone.
 *
 * [formats] is a bitmask of `Barcode.FORMAT_*` — pass `FORMAT_QR_CODE` for member cards and the
 * common retail 1D formats for product barcodes.
 */
@OptIn(ExperimentalGetImage::class)
@Composable
fun CameraScanner(
    formats: Int,
    onDecoded: (String) -> Unit,
    modifier: Modifier = Modifier,
) {
    val lifecycleOwner = LocalLifecycleOwner.current
    val onDecodedState = rememberUpdatedState(onDecoded)
    // Barcode detection must NOT run on the main thread — a dedicated single worker keeps the
    // preview smooth (STRATEGY_KEEP_ONLY_LATEST drops stale frames while a detection is in flight).
    val analysisExecutor = remember { Executors.newSingleThreadExecutor() }
    val barcodeClient = remember(formats) {
        BarcodeScanning.getClient(BarcodeScannerOptions.Builder().setBarcodeFormats(formats).build())
    }
    val cameraProviderRef = remember { java.util.concurrent.atomic.AtomicReference<ProcessCameraProvider?>() }

    DisposableEffect(Unit) {
        onDispose {
            cameraProviderRef.get()?.unbindAll()
            analysisExecutor.shutdown()
            barcodeClient.close()
        }
    }

    AndroidView(
        modifier = modifier,
        factory = { ctx ->
            val previewView = PreviewView(ctx).apply {
                scaleType = PreviewView.ScaleType.FILL_CENTER
                implementationMode = PreviewView.ImplementationMode.COMPATIBLE
            }
            val cameraProviderFuture = ProcessCameraProvider.getInstance(ctx)
            cameraProviderFuture.addListener({
                val cameraProvider = cameraProviderFuture.get()
                cameraProviderRef.set(cameraProvider)

                val preview = Preview.Builder().build().also {
                    it.setSurfaceProvider(previewView.surfaceProvider)
                }

                val analysis = ImageAnalysis.Builder()
                    .setBackpressureStrategy(ImageAnalysis.STRATEGY_KEEP_ONLY_LATEST)
                    .setTargetResolution(Size(1280, 720))
                    // Convert to RGBA so we decode from a clean Bitmap — the emulated webcam's YUV
                    // planes are often malformed and ML Kit can't decode them. RGBA sidesteps that.
                    .setOutputImageFormat(ImageAnalysis.OUTPUT_IMAGE_FORMAT_RGBA_8888)
                    .build()
                analysis.setAnalyzer(analysisExecutor) { imageProxy ->
                    val input = try {
                        InputImage.fromBitmap(imageProxy.toBitmap(), imageProxy.imageInfo.rotationDegrees)
                    } catch (_: Exception) {
                        imageProxy.close(); return@setAnalyzer
                    }
                    barcodeClient.process(input)
                        .addOnSuccessListener { barcodes ->
                            barcodes.firstOrNull { !it.rawValue.isNullOrBlank() }?.rawValue?.let { value ->
                                onDecodedState.value(value)
                            }
                        }
                        .addOnCompleteListener { imageProxy.close() }
                }

                try {
                    cameraProvider.unbindAll()
                    cameraProvider.bindToLifecycle(
                        lifecycleOwner,
                        CameraSelector.DEFAULT_BACK_CAMERA,
                        preview,
                        analysis,
                    )
                } catch (_: Exception) {
                    // No usable camera (e.g. AVD with camera disabled) — caller shows the manual fallback.
                }
            }, ContextCompat.getMainExecutor(ctx))
            previewView
        },
    )
}

/** All the barcode symbologies we expect on product labels, plus QR. */
const val PRODUCT_BARCODE_FORMATS: Int =
    Barcode.FORMAT_EAN_13 or Barcode.FORMAT_EAN_8 or Barcode.FORMAT_UPC_A or
        Barcode.FORMAT_UPC_E or Barcode.FORMAT_CODE_128 or Barcode.FORMAT_CODE_39 or
        Barcode.FORMAT_QR_CODE

const val MEMBER_QR_FORMATS: Int = Barcode.FORMAT_QR_CODE

/**
 * Remembers whether the CAMERA permission is currently granted, and returns a requester lambda.
 * Compose-friendly wrapper around the Activity Result API so screens can gate the [CameraScanner].
 */
@Composable
fun rememberCameraPermission(): Pair<Boolean, () -> Unit> {
    val context = LocalContext.current
    var granted by remember {
        mutableStateOf(
            ContextCompat.checkSelfPermission(context, Manifest.permission.CAMERA) ==
                PackageManager.PERMISSION_GRANTED,
        )
    }
    val launcher = rememberLauncherForActivityResult(
        ActivityResultContracts.RequestPermission(),
    ) { isGranted -> granted = isGranted }
    return granted to { launcher.launch(Manifest.permission.CAMERA) }
}
