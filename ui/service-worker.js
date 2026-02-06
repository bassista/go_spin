self.addEventListener('install', event => {
  event.waitUntil(
    caches.open('gosspin-ui-v1').then(cache => {
      return cache.addAll([
        './index.html',
        './assets/app.js',
        './assets/app-icon-192.png',
        './assets/app-icon-512.png'
      ]);
    })
  );
});

self.addEventListener('fetch', event => {
  event.respondWith(
    caches.match(event.request).then(response => {
      return response || fetch(event.request);
    })
  );
});
