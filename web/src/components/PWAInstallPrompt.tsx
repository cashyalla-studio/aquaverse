import { useState, useEffect } from 'react'

interface BeforeInstallPromptEvent extends Event {
  prompt(): Promise<void>
  userChoice: Promise<{ outcome: 'accepted' | 'dismissed' }>
}

export function PWAInstallPrompt() {
  const [deferredPrompt, setDeferredPrompt] = useState<BeforeInstallPromptEvent | null>(null)
  const [showPrompt, setShowPrompt] = useState(false)

  useEffect(() => {
    const handler = (e: Event) => {
      e.preventDefault()
      setDeferredPrompt(e as BeforeInstallPromptEvent)
      // 5초 후 프롬프트 표시
      setTimeout(() => setShowPrompt(true), 5000)
    }
    window.addEventListener('beforeinstallprompt', handler)
    return () => window.removeEventListener('beforeinstallprompt', handler)
  }, [])

  const handleInstall = async () => {
    if (!deferredPrompt) return
    await deferredPrompt.prompt()
    const choice = await deferredPrompt.userChoice
    if (choice.outcome === 'accepted') {
      setDeferredPrompt(null)
      setShowPrompt(false)
    }
  }

  if (!showPrompt || !deferredPrompt) return null

  return (
    <div className="fixed bottom-20 left-4 right-4 z-50 bg-white rounded-2xl shadow-xl border border-gray-100 p-4 flex items-center gap-3">
      <div className="w-10 h-10 bg-sky-500 rounded-xl flex items-center justify-center text-white text-xl flex-shrink-0">
        🐠
      </div>
      <div className="flex-1 min-w-0">
        <p className="font-semibold text-gray-900 text-sm">AquaVerse 설치</p>
        <p className="text-xs text-gray-500">홈 화면에 추가하고 앱처럼 사용하세요</p>
      </div>
      <div className="flex gap-2 flex-shrink-0">
        <button
          onClick={() => setShowPrompt(false)}
          className="px-3 py-1.5 text-sm text-gray-500 hover:text-gray-700"
        >
          나중에
        </button>
        <button
          onClick={handleInstall}
          className="px-3 py-1.5 text-sm bg-sky-500 text-white rounded-lg font-medium hover:bg-sky-600"
        >
          설치
        </button>
      </div>
    </div>
  )
}
