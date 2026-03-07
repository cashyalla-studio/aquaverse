class ListingDetail {
  final int id;
  final String title;
  final String? description;
  final String fishScientificName;
  final String fishCommonName;
  final String priceKrw;
  final bool isFree;
  final String healthStatus;
  final String tradeType;
  final String location;
  final String sellerNickname;
  final double sellerTrustScore;
  final List<String> imageUrls;
  final String status;
  final DateTime createdAt;
  final int? quantity;
  final String? sex;
  final double? currentSizeCm;
  final int? ageMonths;
  final bool bredBySeller;

  const ListingDetail({
    required this.id,
    required this.title,
    this.description,
    required this.fishScientificName,
    required this.fishCommonName,
    required this.priceKrw,
    required this.isFree,
    required this.healthStatus,
    required this.tradeType,
    required this.location,
    required this.sellerNickname,
    required this.sellerTrustScore,
    required this.imageUrls,
    required this.status,
    required this.createdAt,
    this.quantity,
    this.sex,
    this.currentSizeCm,
    this.ageMonths,
    required this.bredBySeller,
  });

  factory ListingDetail.fromJson(Map<String, dynamic> json) => ListingDetail(
    id: json['id'],
    title: json['title'] ?? '',
    description: json['description'],
    fishScientificName: json['fish_scientific_name'] ?? '',
    fishCommonName: json['fish_common_name'] ?? '',
    priceKrw: json['price_krw']?.toString() ?? '0',
    isFree: json['is_free'] ?? false,
    healthStatus: json['health_status'] ?? 'GOOD',
    tradeType: json['trade_type'] ?? 'DIRECT',
    location: json['location'] ?? '',
    sellerNickname: json['seller_nickname'] ?? '',
    sellerTrustScore: (json['seller_trust_score'] as num?)?.toDouble() ?? 36.5,
    imageUrls: List<String>.from(json['image_urls'] ?? []),
    status: json['status'] ?? 'ACTIVE',
    createdAt: DateTime.tryParse(json['created_at'] ?? '') ?? DateTime.now(),
    quantity: json['quantity'],
    sex: json['sex'],
    currentSizeCm: (json['current_size_cm'] as num?)?.toDouble(),
    ageMonths: json['age_months'],
    bredBySeller: json['bred_by_seller'] ?? false,
  );
}

class CreateListingRequest {
  final String title;
  final String? description;
  final int fishDataId;
  final String priceKrw;
  final String healthStatus;
  final String tradeType;
  final String location;
  final double? latitude;
  final double? longitude;
  final int? quantity;
  final String? sex;
  final double? currentSizeCm;
  final int? ageMonths;
  final bool bredBySeller;

  const CreateListingRequest({
    required this.title,
    this.description,
    required this.fishDataId,
    required this.priceKrw,
    required this.healthStatus,
    required this.tradeType,
    required this.location,
    this.latitude,
    this.longitude,
    this.quantity,
    this.sex,
    this.currentSizeCm,
    this.ageMonths,
    required this.bredBySeller,
  });

  Map<String, dynamic> toJson() => {
    'title': title,
    if (description != null) 'description': description,
    'fish_data_id': fishDataId,
    'price_krw': priceKrw,
    'health_status': healthStatus,
    'trade_type': tradeType,
    'location': location,
    if (latitude != null) 'latitude': latitude,
    if (longitude != null) 'longitude': longitude,
    if (quantity != null) 'quantity': quantity,
    if (sex != null) 'sex': sex,
    if (currentSizeCm != null) 'current_size_cm': currentSizeCm,
    if (ageMonths != null) 'age_months': ageMonths,
    'bred_by_seller': bredBySeller,
  };
}
