// PostHog analytics wrapper
// 실제 프로덕션에서는 환경변수로 키 설정
const POSTHOG_KEY = import.meta.env.VITE_POSTHOG_KEY || '';
const POSTHOG_HOST = import.meta.env.VITE_POSTHOG_HOST || 'https://app.posthog.com';

let posthogLoaded = false;

export async function initPostHog() {
  if (!POSTHOG_KEY || POSTHOG_KEY === 'phc_dev_local_placeholder') return;
  try {
    const { default: posthog } = await import('posthog-js');
    posthog.init(POSTHOG_KEY, {
      api_host: POSTHOG_HOST,
      capture_pageview: false, // 수동으로 관리
      persistence: 'localStorage',
    });
    posthogLoaded = true;
  } catch (_) {}
}

export function capture(event: string, props?: Record<string, unknown>) {
  if (!posthogLoaded) return;
  import('posthog-js').then(({ default: ph }) => ph.capture(event, props)).catch(() => {});
}

// 핵심 5가지 이벤트
export const events = {
  speciesViewed: (id: number, name: string) =>
    capture('species_viewed', { species_id: id, species_name: name }),
  listingCreated: (fishName: string, price: number) =>
    capture('listing_created', { fish_name: fishName, price }),
  tradeInitiated: (listingId: number) =>
    capture('trade_initiated', { listing_id: listingId }),
  searchPerformed: (query: string, resultCount: number) =>
    capture('search_performed', { query, result_count: resultCount }),
  userRegistered: (locale: string) =>
    capture('user_registered', { locale }),
};
