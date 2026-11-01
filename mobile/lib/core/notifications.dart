import 'package:flutter_local_notifications/flutter_local_notifications.dart';

final _plugin = FlutterLocalNotificationsPlugin();
Future<bool?>? _ready;

Future<void> _init() => _ready ??= _plugin.initialize(
  settings: const InitializationSettings(
    android: AndroidInitializationSettings('@mipmap/ic_launcher'),
    iOS: DarwinInitializationSettings(
      requestAlertPermission: false,
      requestBadgePermission: false,
      requestSoundPermission: false,
    ),
  ),
);

Future<void> requestNotificationPermission() async {
  await _init();
  await _plugin
      .resolvePlatformSpecificImplementation<AndroidFlutterLocalNotificationsPlugin>()
      ?.requestNotificationsPermission();
  await _plugin.resolvePlatformSpecificImplementation<IOSFlutterLocalNotificationsPlugin>()?.requestPermissions(
    alert: true,
    sound: true,
  );
}

/// Tells the BA a queued order went through or failed, even from another screen or the background.
/// Silently skipped when notifications are not allowed.
Future<void> notifyOrder(String orderKey, String title, String body) async {
  await _init();
  await _plugin.show(
    id: orderKey.hashCode,
    title: title,
    body: body,
    notificationDetails: const NotificationDetails(
      android: AndroidNotificationDetails('order_updates', 'Order updates'),
      iOS: DarwinNotificationDetails(),
    ),
  );
}
