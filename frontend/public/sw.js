const CACHE = 'sunpath-v1'

const PRECACHE_URLS = [
  '/',
  '/index.html',
  '/manifest.json',
]

self.addEventListener('install', (event) => {
  event.waitUntil(
    caches.open(CACHE).then((cache) => cache.addAll(PRECACHE_URLS))
  )
})

self.addEventListener('activate', (event) => {
  event.waitUntil(
    caches.keys().then((names) =>
      Promise.all(names.filter((n) => n !== CACHE).map((n) => caches.delete(n)))
    )
  )
})

self.addEventListener('fetch', (event) => {
  const url = new URL(event.request.url)

  if (url.protocol !== 'http:' && url.protocol !== 'https:') {
    return
  }

  if (url.pathname.startsWith('/api/')) {
    return
  }

  event.respondWith(
    caches.match(event.request).then((cached) => {
      const fetched = fetch(event.request).then((response) => {
        if (response.ok && response.type === 'basic') {
          const copy = response.clone()
          caches.open(CACHE).then((cache) => cache.put(event.request, copy))
        }
        return response
      }).catch(() => cached)
      return cached || fetched
    })
  )
})
