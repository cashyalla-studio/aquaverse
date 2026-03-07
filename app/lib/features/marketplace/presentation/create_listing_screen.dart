import 'package:flutter/material.dart';
import 'package:flutter/services.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:go_router/go_router.dart';
import '../../../l10n/app_localizations.dart';
import '../../fish/data/fish_repository.dart';
import '../../fish/domain/fish_model.dart';
import '../data/listing_repository.dart';
import '../domain/listing_model.dart';
import 'cites_warning_widget.dart';

class CreateListingScreen extends ConsumerStatefulWidget {
  const CreateListingScreen({super.key});

  @override
  ConsumerState<CreateListingScreen> createState() => _CreateListingScreenState();
}

class _CreateListingScreenState extends ConsumerState<CreateListingScreen> {
  final _pageController = PageController();
  int _currentStep = 0;
  bool _submitting = false;

  // 폼 데이터
  FishListItem? _selectedFish;
  final _titleController = TextEditingController();
  final _descriptionController = TextEditingController();
  final _priceController = TextEditingController();
  bool _isFree = false;
  String _healthStatus = 'GOOD';
  String _tradeType = 'DIRECT';
  final _locationController = TextEditingController();
  int? _quantity;
  String? _sex;
  double? _currentSizeCm;
  int? _ageMonths;
  bool _bredBySeller = false;

  @override
  void dispose() {
    _pageController.dispose();
    _titleController.dispose();
    _descriptionController.dispose();
    _priceController.dispose();
    _locationController.dispose();
    super.dispose();
  }

  void _nextStep() {
    if (_currentStep < 2) {
      setState(() => _currentStep++);
      _pageController.nextPage(duration: const Duration(milliseconds: 300), curve: Curves.easeInOut);
    }
  }

  void _prevStep() {
    if (_currentStep > 0) {
      setState(() => _currentStep--);
      _pageController.previousPage(duration: const Duration(milliseconds: 300), curve: Curves.easeInOut);
    }
  }

  bool get _step0Valid => _selectedFish != null;
  bool get _step1Valid =>
    _titleController.text.trim().isNotEmpty &&
    _locationController.text.trim().isNotEmpty &&
    (_isFree || _priceController.text.trim().isNotEmpty);

  Future<void> _submit() async {
    if (!_step1Valid || _submitting) return;
    setState(() => _submitting = true);

    final locale = Localizations.localeOf(context).toString();
    final repo = ref.read(listingRepositoryProvider(locale));

    try {
      final req = CreateListingRequest(
        title: _titleController.text.trim(),
        description: _descriptionController.text.trim().isEmpty ? null : _descriptionController.text.trim(),
        fishDataId: _selectedFish!.id,
        priceKrw: _isFree ? '0' : _priceController.text.replaceAll(',', '').trim(),
        healthStatus: _healthStatus,
        tradeType: _tradeType,
        location: _locationController.text.trim(),
        quantity: _quantity,
        sex: _sex,
        currentSizeCm: _currentSizeCm,
        ageMonths: _ageMonths,
        bredBySeller: _bredBySeller,
      );

      final id = await repo.createListing(req);
      if (mounted) {
        context.go('/marketplace/$id');
      }
    } catch (_) {
      setState(() => _submitting = false);
      if (mounted) {
        ScaffoldMessenger.of(context).showSnackBar(
          const SnackBar(content: Text('Failed to create listing. Please try again.')),
        );
      }
    }
  }

  @override
  Widget build(BuildContext context) {
    final l10n = AppLocalizations.of(context)!;

    return Scaffold(
      appBar: AppBar(
        title: Text(l10n.marketplaceCreateListing),
        backgroundColor: const Color(0xFF0EA5E9),
        foregroundColor: Colors.white,
        leading: IconButton(
          icon: const Icon(Icons.close),
          onPressed: () => context.pop(),
        ),
      ),
      body: Column(
        children: [
          // 스텝 인디케이터
          Padding(
            padding: const EdgeInsets.symmetric(horizontal: 24, vertical: 12),
            child: Row(
              children: List.generate(3, (i) => Expanded(
                child: Padding(
                  padding: EdgeInsets.only(right: i < 2 ? 6 : 0),
                  child: AnimatedContainer(
                    duration: const Duration(milliseconds: 200),
                    height: 4,
                    decoration: BoxDecoration(
                      color: i <= _currentStep ? const Color(0xFF0EA5E9) : const Color(0xFFE5E7EB),
                      borderRadius: BorderRadius.circular(2),
                    ),
                  ),
                ),
              )),
            ),
          ),

          Expanded(
            child: PageView(
              controller: _pageController,
              physics: const NeverScrollableScrollPhysics(),
              children: [
                _Step1FishSelect(
                  selected: _selectedFish,
                  onSelect: (fish) => setState(() => _selectedFish = fish),
                  locale: Localizations.localeOf(context).toString(),
                  l10n: l10n,
                ),
                _Step2Details(
                  l10n: l10n,
                  titleController: _titleController,
                  descriptionController: _descriptionController,
                  priceController: _priceController,
                  isFree: _isFree,
                  onFreeToggle: (v) => setState(() => _isFree = v),
                  healthStatus: _healthStatus,
                  onHealthChange: (v) => setState(() => _healthStatus = v),
                  tradeType: _tradeType,
                  onTradeChange: (v) => setState(() => _tradeType = v),
                  locationController: _locationController,
                ),
                _Step3Extra(
                  l10n: l10n,
                  quantity: _quantity,
                  onQuantityChange: (v) => setState(() => _quantity = v),
                  sex: _sex,
                  onSexChange: (v) => setState(() => _sex = v),
                  currentSizeCm: _currentSizeCm,
                  onSizeChange: (v) => setState(() => _currentSizeCm = v),
                  ageMonths: _ageMonths,
                  onAgeChange: (v) => setState(() => _ageMonths = v),
                  bredBySeller: _bredBySeller,
                  onBredChange: (v) => setState(() => _bredBySeller = v),
                ),
              ],
            ),
          ),

          // 네비게이션 버튼
          SafeArea(
            child: Padding(
              padding: const EdgeInsets.all(16),
              child: Row(
                children: [
                  if (_currentStep > 0) ...[
                    OutlinedButton(
                      onPressed: _prevStep,
                      child: Text(l10n.commonCancel),
                    ),
                    const SizedBox(width: 12),
                  ],
                  Expanded(
                    child: ElevatedButton(
                      onPressed: _currentStep == 0 ? (_step0Valid ? _nextStep : null)
                          : _currentStep == 1 ? (_step1Valid ? _nextStep : null)
                          : (_submitting ? null : _submit),
                      style: ElevatedButton.styleFrom(
                        backgroundColor: const Color(0xFF0EA5E9),
                        foregroundColor: Colors.white,
                        minimumSize: const Size.fromHeight(48),
                        shape: RoundedRectangleBorder(borderRadius: BorderRadius.circular(12)),
                      ),
                      child: _submitting
                        ? const SizedBox(width: 20, height: 20, child: CircularProgressIndicator(strokeWidth: 2, color: Colors.white))
                        : Text(_currentStep < 2 ? 'Next' : 'Post Listing', style: const TextStyle(fontWeight: FontWeight.bold)),
                    ),
                  ),
                ],
              ),
            ),
          ),
        ],
      ),
    );
  }
}

// 스텝 1: 어종 선택
class _Step1FishSelect extends ConsumerStatefulWidget {
  final FishListItem? selected;
  final ValueChanged<FishListItem> onSelect;
  final String locale;
  final AppLocalizations l10n;
  const _Step1FishSelect({
    required this.selected, required this.onSelect,
    required this.locale, required this.l10n,
  });

  @override
  ConsumerState<_Step1FishSelect> createState() => _Step1FishSelectState();
}

class _Step1FishSelectState extends ConsumerState<_Step1FishSelect> {
  String _query = '';

  @override
  Widget build(BuildContext context) {
    final params = {
      'locale': widget.locale,
      'query': _query,
      'care_level': '',
      'page': 1,
    };
    final fishAsync = ref.watch(fishListProvider(params));

    return Column(
      crossAxisAlignment: CrossAxisAlignment.start,
      children: [
        Padding(
          padding: const EdgeInsets.fromLTRB(16, 8, 16, 8),
          child: Column(
            crossAxisAlignment: CrossAxisAlignment.start,
            children: [
              const Text('Which fish are you listing?',
                  style: TextStyle(fontSize: 17, fontWeight: FontWeight.bold)),
              const SizedBox(height: 12),
              TextField(
                onChanged: (v) => setState(() => _query = v),
                decoration: InputDecoration(
                  hintText: widget.l10n.commonSearch,
                  prefixIcon: const Icon(Icons.search),
                  border: OutlineInputBorder(borderRadius: BorderRadius.circular(12)),
                  contentPadding: const EdgeInsets.symmetric(vertical: 0),
                ),
              ),
              if (widget.selected != null) ...[
                const SizedBox(height: 8),
                Container(
                  padding: const EdgeInsets.all(10),
                  decoration: BoxDecoration(
                    color: const Color(0xFFE0F2FE),
                    borderRadius: BorderRadius.circular(10),
                  ),
                  child: Row(
                    children: [
                      const Icon(Icons.check_circle, color: Color(0xFF0EA5E9), size: 18),
                      const SizedBox(width: 8),
                      Expanded(child: Text(widget.selected!.commonName, style: const TextStyle(fontWeight: FontWeight.w600))),
                    ],
                  ),
                ),
                CitesWarningWidget(scientificName: widget.selected!.scientificName),
              ],
            ],
          ),
        ),
        Expanded(
          child: fishAsync.when(
            loading: () => const Center(child: CircularProgressIndicator()),
            error: (_, __) => const Center(child: Text('Failed to load fish')),
            data: (result) => ListView.builder(
              padding: const EdgeInsets.symmetric(horizontal: 16),
              itemCount: result.items.length,
              itemBuilder: (context, i) {
                final fish = result.items[i];
                final isSelected = widget.selected?.id == fish.id;
                return ListTile(
                  onTap: () => widget.onSelect(fish),
                  leading: CircleAvatar(
                    backgroundColor: isSelected ? const Color(0xFF0EA5E9) : const Color(0xFFE0F2FE),
                    child: Text(isSelected ? '✓' : '🐠'),
                  ),
                  title: Text(fish.commonName, style: const TextStyle(fontWeight: FontWeight.w600)),
                  subtitle: Text(fish.scientificName, style: const TextStyle(fontStyle: FontStyle.italic)),
                  trailing: isSelected ? const Icon(Icons.check_circle, color: Color(0xFF0EA5E9)) : null,
                  shape: RoundedRectangleBorder(borderRadius: BorderRadius.circular(10)),
                  tileColor: isSelected ? const Color(0xFFE0F2FE) : null,
                );
              },
            ),
          ),
        ),
      ],
    );
  }
}

// 스텝 2: 상세 정보
class _Step2Details extends StatelessWidget {
  final AppLocalizations l10n;
  final TextEditingController titleController;
  final TextEditingController descriptionController;
  final TextEditingController priceController;
  final bool isFree;
  final ValueChanged<bool> onFreeToggle;
  final String healthStatus;
  final ValueChanged<String> onHealthChange;
  final String tradeType;
  final ValueChanged<String> onTradeChange;
  final TextEditingController locationController;

  const _Step2Details({
    required this.l10n,
    required this.titleController, required this.descriptionController,
    required this.priceController, required this.isFree, required this.onFreeToggle,
    required this.healthStatus, required this.onHealthChange,
    required this.tradeType, required this.onTradeChange, required this.locationController,
  });

  @override
  Widget build(BuildContext context) {
    return SingleChildScrollView(
      padding: const EdgeInsets.all(16),
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          const Text('Listing Details', style: TextStyle(fontSize: 17, fontWeight: FontWeight.bold)),
          const SizedBox(height: 16),

          _FieldLabel('Title *'),
          TextField(
            controller: titleController,
            decoration: _inputDeco('e.g. Breeding pair of Discus'),
          ),
          const SizedBox(height: 12),

          _FieldLabel('Description'),
          TextField(
            controller: descriptionController,
            decoration: _inputDeco('Describe the fish, tank history, diet...'),
            maxLines: 4,
          ),
          const SizedBox(height: 12),

          _FieldLabel(l10n.marketplacePrice + ' *'),
          Row(
            children: [
              Expanded(
                child: TextField(
                  controller: priceController,
                  enabled: !isFree,
                  keyboardType: TextInputType.number,
                  decoration: _inputDeco('e.g. 50000'),
                ),
              ),
              const SizedBox(width: 12),
              Row(
                children: [
                  Checkbox(value: isFree, onChanged: (v) => onFreeToggle(v ?? false)),
                  Text(l10n.marketplaceFree),
                ],
              ),
            ],
          ),
          const SizedBox(height: 12),

          _FieldLabel(l10n.marketplaceHealthStatus + ' *'),
          _RadioGroup<String>(
            options: const [
              ('EXCELLENT', 'Excellent 💚'),
              ('GOOD', 'Good 💙'),
              ('DISEASE_HISTORY', 'Disease History 💛'),
              ('UNDER_TREATMENT', 'Under Treatment ❤️'),
            ],
            selected: healthStatus,
            onChanged: onHealthChange,
          ),
          const SizedBox(height: 12),

          _FieldLabel(l10n.marketplaceTradeType + ' *'),
          _RadioGroup<String>(
            options: [
              ('DIRECT', '🤝 ${l10n.marketplaceDirect}'),
              ('COURIER', '📦 ${l10n.marketplaceCourier}'),
              ('AQUA_COURIER', '🐟 ${l10n.marketplaceAquaCourier}'),
              ('ALL', '✅ All Methods'),
            ],
            selected: tradeType,
            onChanged: onTradeChange,
          ),
          const SizedBox(height: 12),

          _FieldLabel('Location *'),
          TextField(
            controller: locationController,
            decoration: _inputDeco('e.g. Seoul, Gangnam-gu'),
          ),
          const SizedBox(height: 24),
        ],
      ),
    );
  }

  InputDecoration _inputDeco(String hint) => InputDecoration(
    hintText: hint,
    border: OutlineInputBorder(borderRadius: BorderRadius.circular(10)),
    contentPadding: const EdgeInsets.symmetric(horizontal: 12, vertical: 10),
  );
}

class _FieldLabel extends StatelessWidget {
  final String text;
  const _FieldLabel(this.text);
  @override
  Widget build(BuildContext context) => Padding(
    padding: const EdgeInsets.only(bottom: 6),
    child: Text(text, style: const TextStyle(fontSize: 13, fontWeight: FontWeight.w600, color: Color(0xFF374151))),
  );
}

class _RadioGroup<T> extends StatelessWidget {
  final List<(T, String)> options;
  final T selected;
  final ValueChanged<T> onChanged;
  const _RadioGroup({required this.options, required this.selected, required this.onChanged});

  @override
  Widget build(BuildContext context) => Column(
    children: options.map((opt) => RadioListTile<T>(
      title: Text(opt.$2, style: const TextStyle(fontSize: 13)),
      value: opt.$1,
      groupValue: selected,
      onChanged: (v) { if (v != null) onChanged(v); },
      dense: true,
      contentPadding: EdgeInsets.zero,
    )).toList(),
  );
}

// 스텝 3: 추가 정보
class _Step3Extra extends StatelessWidget {
  final AppLocalizations l10n;
  final int? quantity;
  final ValueChanged<int?> onQuantityChange;
  final String? sex;
  final ValueChanged<String?> onSexChange;
  final double? currentSizeCm;
  final ValueChanged<double?> onSizeChange;
  final int? ageMonths;
  final ValueChanged<int?> onAgeChange;
  final bool bredBySeller;
  final ValueChanged<bool> onBredChange;

  const _Step3Extra({
    required this.l10n,
    required this.quantity, required this.onQuantityChange,
    required this.sex, required this.onSexChange,
    required this.currentSizeCm, required this.onSizeChange,
    required this.ageMonths, required this.onAgeChange,
    required this.bredBySeller, required this.onBredChange,
  });

  @override
  Widget build(BuildContext context) {
    return SingleChildScrollView(
      padding: const EdgeInsets.all(16),
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          const Text('Additional Info (Optional)', style: TextStyle(fontSize: 17, fontWeight: FontWeight.bold)),
          const SizedBox(height: 4),
          const Text('These details help buyers make better decisions.', style: TextStyle(fontSize: 13, color: Color(0xFF6B7280))),
          const SizedBox(height: 20),

          Row(
            children: [
              Expanded(child: _NumberField(
                label: l10n.marketplaceQuantity,
                value: quantity,
                onChanged: onQuantityChange,
                isInt: true,
              )),
              const SizedBox(width: 12),
              Expanded(child: _NumberField(
                label: '${l10n.marketplaceSize} (cm)',
                value: currentSizeCm,
                onChanged: (v) => onSizeChange(v),
                isInt: false,
              )),
            ],
          ),
          const SizedBox(height: 12),

          _NumberField(
            label: '${l10n.marketplaceAge} (months)',
            value: ageMonths,
            onChanged: onAgeChange,
            isInt: true,
          ),
          const SizedBox(height: 12),

          _FieldLabel(l10n.marketplaceSex),
          Wrap(
            spacing: 8,
            children: [
              for (final s in [null, 'MALE', 'FEMALE', 'MIXED', 'UNKNOWN'])
                ChoiceChip(
                  label: Text(s == null ? 'Unknown' : s[0] + s.substring(1).toLowerCase()),
                  selected: sex == s,
                  onSelected: (_) => onSexChange(s),
                  selectedColor: const Color(0xFFE0F2FE),
                ),
            ],
          ),
          const SizedBox(height: 16),

          SwitchListTile(
            title: Text(l10n.marketplaceBredBySeller),
            subtitle: const Text('Fish were bred in your own tank', style: TextStyle(fontSize: 12)),
            value: bredBySeller,
            onChanged: onBredChange,
            activeColor: const Color(0xFF0EA5E9),
            contentPadding: EdgeInsets.zero,
          ),
        ],
      ),
    );
  }
}

class _FieldLabel extends StatelessWidget {
  final String text;
  const _FieldLabel(this.text);
  @override
  Widget build(BuildContext context) => Padding(
    padding: const EdgeInsets.only(bottom: 6),
    child: Text(text, style: const TextStyle(fontSize: 13, fontWeight: FontWeight.w600, color: Color(0xFF374151))),
  );
}

class _NumberField<T extends num?> extends StatelessWidget {
  final String label;
  final T value;
  final ValueChanged<T> onChanged;
  final bool isInt;
  const _NumberField({required this.label, required this.value, required this.onChanged, required this.isInt});

  @override
  Widget build(BuildContext context) {
    return Column(
      crossAxisAlignment: CrossAxisAlignment.start,
      children: [
        _FieldLabel(label),
        TextField(
          controller: TextEditingController(text: value?.toString() ?? ''),
          keyboardType: isInt ? TextInputType.number : const TextInputType.numberWithOptions(decimal: true),
          inputFormatters: [
            isInt ? FilteringTextInputFormatter.digitsOnly : FilteringTextInputFormatter.allow(RegExp(r'[0-9.]')),
          ],
          decoration: InputDecoration(
            border: OutlineInputBorder(borderRadius: BorderRadius.circular(10)),
            contentPadding: const EdgeInsets.symmetric(horizontal: 12, vertical: 10),
          ),
          onChanged: (v) {
            if (v.isEmpty) {
              onChanged(null as T);
            } else if (isInt) {
              onChanged(int.tryParse(v) as T);
            } else {
              onChanged(double.tryParse(v) as T);
            }
          },
        ),
      ],
    );
  }
}
