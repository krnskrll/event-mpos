/// Override per build: `flutter run --dart-define=API_BASE_URL=http://192.168.1.10:8080/`.
/// The default reaches the host machine from the Android emulator.
const apiBaseUrl = String.fromEnvironment('API_BASE_URL', defaultValue: 'http://10.0.2.2:8080/');
