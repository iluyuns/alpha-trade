import { initializeApp } from "https://www.gstatic.com/firebasejs/12.8.0/firebase-app.js";
import { getMessaging, onBackgroundMessage } from "https://www.gstatic.com/firebasejs/12.8.0/firebase-messaging.js";

const firebaseConfig = {
  apiKey: "AIzaSyD9JVLzirRWwR4y1c1fFsxXD11YkJOGb0w",
  authDomain: "coinpulse-666.firebaseapp.com",
  projectId: "coinpulse-666",
  storageBucket: "coinpulse-666.firebasestorage.app",
  messagingSenderId: "1053322751533",
  appId: "1:1053322751533:web:dd7c91f6116a0de90012a4"
};

const app = initializeApp(firebaseConfig);
const messaging = getMessaging(app);

// Handle background messages
onBackgroundMessage(messaging, (payload) => {
  console.log('[firebase-messaging-sw.js] Received background message ', payload);
  const notificationTitle = payload.notification.title;
  const notificationOptions = {
    body: payload.notification.body,
    icon: '/firebase-logo.png'
  };

  self.registration.showNotification(notificationTitle, notificationOptions);
});
