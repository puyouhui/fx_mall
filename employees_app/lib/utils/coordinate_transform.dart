import 'package:latlong2/latlong.dart';
import 'dart:math' as math;

/// 坐标转换工具：GCJ-02（高德地图）与 WGS84（GPS）之间的转换
/// 天地图使用 WGS84 坐标系，高德地图使用 GCJ-02 坐标系

class CoordinateTransform {
  // 圆周率
  static const double pi = 3.1415926535897932384626;
  // 长半轴
  static const double a = 6378245.0;
  // 扁率
  static const double ee = 0.00669342162296594323;

  /// GCJ-02 转 WGS84
  /// 将高德地图坐标转换为GPS坐标（天地图使用）
  static LatLng gcj02ToWgs84(double lat, double lng) {
    double dLat = _transformLat(lng - 105.0, lat - 35.0);
    double dLng = _transformLng(lng - 105.0, lat - 35.0);
    double radLat = lat / 180.0 * pi;
    double magic = math.sin(radLat);
    magic = 1 - ee * magic * magic;
    double sqrtMagic = math.sqrt(magic);
    dLat = (dLat * 180.0) / ((a * (1 - ee)) / (magic * sqrtMagic) * pi);
    dLng = (dLng * 180.0) / (a / sqrtMagic * math.cos(radLat) * pi);
    double mgLat = lat + dLat;
    double mgLng = lng + dLng;
    return LatLng(lat * 2 - mgLat, lng * 2 - mgLng);
  }

  /// WGS84 转 GCJ-02
  /// 将GPS坐标转换为高德地图坐标
  static LatLng wgs84ToGcj02(double lat, double lng) {
    double dLat = _transformLat(lng - 105.0, lat - 35.0);
    double dLng = _transformLng(lng - 105.0, lat - 35.0);
    double radLat = lat / 180.0 * pi;
    double magic = math.sin(radLat);
    magic = 1 - ee * magic * magic;
    double sqrtMagic = math.sqrt(magic);
    dLat = (dLat * 180.0) / ((a * (1 - ee)) / (magic * sqrtMagic) * pi);
    dLng = (dLng * 180.0) / (a / sqrtMagic * math.cos(radLat) * pi);
    double mgLat = lat + dLat;
    double mgLng = lng + dLng;
    return LatLng(mgLat, mgLng);
  }

  static double _transformLat(double lng, double lat) {
    double ret =
        -100.0 +
        2.0 * lng +
        3.0 * lat +
        0.2 * lat * lat +
        0.1 * lng * lat +
        0.2 * math.sqrt(lng.abs());
    ret +=
        (20.0 * math.sin(6.0 * lng * pi) + 20.0 * math.sin(2.0 * lng * pi)) *
        2.0 /
        3.0;
    ret +=
        (20.0 * math.sin(lat * pi) + 40.0 * math.sin(lat / 3.0 * pi)) *
        2.0 /
        3.0;
    ret +=
        (160.0 * math.sin(lat / 12.0 * pi) + 320 * math.sin(lat * pi / 30.0)) *
        2.0 /
        3.0;
    return ret;
  }

  static double _transformLng(double lng, double lat) {
    double ret =
        300.0 +
        lng +
        2.0 * lat +
        0.1 * lng * lng +
        0.1 * lng * lat +
        0.1 * math.sqrt(lng.abs());
    ret +=
        (20.0 * math.sin(6.0 * lng * pi) + 20.0 * math.sin(2.0 * lng * pi)) *
        2.0 /
        3.0;
    ret +=
        (20.0 * math.sin(lng * pi) + 40.0 * math.sin(lng / 3.0 * pi)) *
        2.0 /
        3.0;
    ret +=
        (150.0 * math.sin(lng / 12.0 * pi) +
            300.0 * math.sin(lng / 30.0 * pi)) *
        2.0 /
        3.0;
    return ret;
  }
}
