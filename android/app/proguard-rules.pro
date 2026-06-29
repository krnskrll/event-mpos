# Keep kotlinx.serialization generated serializers.
-keepattributes *Annotation*, InnerClasses
-dontnote kotlinx.serialization.**
-keepclassmembers class com.eventmpos.mpos.data.** {
    *** Companion;
}
-keepclasseswithmembers class com.eventmpos.mpos.data.** {
    kotlinx.serialization.KSerializer serializer(...);
}
