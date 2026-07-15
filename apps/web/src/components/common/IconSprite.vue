<template>
  <svg class="icon-sprite" aria-hidden="true">
    <symbol id="i-grid" viewBox="0 0 24 24"><path d="M4 4h6v6H4V4Zm10 0h6v6h-6V4ZM4 14h6v6H4v-6Zm10 0h6v6h-6v-6Z"/></symbol>
    <symbol id="i-play" viewBox="0 0 24 24"><path d="m8 5 11 7-11 7V5Z"/></symbol>
    <symbol id="i-board" viewBox="0 0 24 24"><path d="M5 3h14a2 2 0 0 1 2 2v14a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2V5a2 2 0 0 1 2-2Zm2 0v18M17 3v18M3 8h18M3 16h18"/></symbol>
    <symbol id="i-record" viewBox="0 0 24 24"><path d="M6 3h9l4 4v14H6a2 2 0 0 1-2-2V5a2 2 0 0 1 2-2Zm8 0v5h5M8 12h8M8 16h6"/></symbol>
    <symbol id="i-spark" viewBox="0 0 24 24"><path d="m12 2 1.7 5.3L19 9l-5.3 1.7L12 16l-1.7-5.3L5 9l5.3-1.7L12 2Zm6 13 .9 2.6 2.6.9-2.6.9L18 22l-.9-2.6-2.6-.9 2.6-.9L18 15Z"/></symbol>
    <symbol id="i-chart" viewBox="0 0 24 24"><path d="M4 20V10m6 10V4m6 16v-7m4 7H2"/></symbol>
    <symbol id="i-clock" viewBox="0 0 24 24"><circle cx="12" cy="12" r="9"/><path d="M12 7v5l3 2"/></symbol>
    <symbol id="i-settings" viewBox="0 0 24 24"><circle cx="12" cy="12" r="3"/><path d="M19 12a7 7 0 0 0-.1-1.2l2-1.5-2-3.4-2.5 1a8 8 0 0 0-2-1.2L14 3h-4l-.4 2.7a8 8 0 0 0-2 1.2l-2.5-1-2 3.4 2 1.5A7 7 0 0 0 5 12c0 .4 0 .8.1 1.2l-2 1.5 2 3.4 2.5-1a8 8 0 0 0 2 1.2L10 21h4l.4-2.7a8 8 0 0 0 2-1.2l2.5 1 2-3.4-2-1.5c.1-.4.1-.8.1-1.2Z"/></symbol>
    <symbol id="i-menu" viewBox="0 0 24 24"><path d="M4 7h16M4 12h16M4 17h16"/></symbol>
    <symbol id="i-close" viewBox="0 0 24 24"><path d="m6 6 12 12M18 6 6 18"/></symbol>
    <symbol id="i-sun" viewBox="0 0 24 24"><circle cx="12" cy="12" r="4"/><path d="M12 2v2m0 16v2M4.9 4.9l1.4 1.4m11.4 11.4 1.4 1.4M2 12h2m16 0h2M4.9 19.1l1.4-1.4M17.7 6.3l1.4-1.4"/></symbol>
    <symbol id="i-moon" viewBox="0 0 24 24"><path d="M20 15.5A8.5 8.5 0 0 1 8.5 4 8.5 8.5 0 1 0 20 15.5Z"/></symbol>
    <symbol id="i-bell" viewBox="0 0 24 24"><path d="M18 9a6 6 0 0 0-12 0c0 7-3 7-3 7h18s-3 0-3-7ZM10 20h4"/></symbol>
    <symbol id="i-arrow" viewBox="0 0 24 24"><path d="M5 12h14m-5-5 5 5-5 5"/></symbol>
    <symbol id="i-upload" viewBox="0 0 24 24"><path d="M12 16V4m-4 4 4-4 4 4M4 15v4a1 1 0 0 0 1 1h14a1 1 0 0 0 1-1v-4"/></symbol>
    <symbol id="i-check" viewBox="0 0 24 24"><path d="m5 12 4 4L19 6"/></symbol>
    <symbol id="i-chevron" viewBox="0 0 24 24"><path d="m9 6 6 6-6 6"/></symbol>
    <symbol id="i-more" viewBox="0 0 24 24"><circle cx="5" cy="12" r="1.5"/><circle cx="12" cy="12" r="1.5"/><circle cx="19" cy="12" r="1.5"/></symbol>
    <symbol id="i-info" viewBox="0 0 24 24"><circle cx="12" cy="12" r="9"/><path d="M12 11v6M12 7h.01"/></symbol>
    <symbol id="i-refresh" viewBox="0 0 24 24"><path d="M20 11a8 8 0 1 0-2.3 5.7M20 4v7h-7"/></symbol>
    <symbol id="i-flip" viewBox="0 0 24 24"><path d="M7 7h11l-3-3m3 13H7l3 3M5 5v14"/></symbol>
    <symbol id="i-volume" viewBox="0 0 24 24"><path d="M5 9v6h4l5 4V5L9 9H5Zm12 1a3 3 0 0 1 0 4m2-7a7 7 0 0 1 0 10"/></symbol>
    <symbol id="i-flag" viewBox="0 0 24 24"><path d="M5 21V4m0 1h11l-2 3 2 3H5"/></symbol>
    <symbol id="i-undo" viewBox="0 0 24 24"><path d="M9 8 5 12l4 4m-4-4h8a6 6 0 0 1 6 6"/></symbol>
    <symbol id="i-eye" viewBox="0 0 24 24"><path d="M2.5 12s3.5-6 9.5-6 9.5 6 9.5 6-3.5 6-9.5 6-9.5-6-9.5-6Z"/><circle cx="12" cy="12" r="2.5"/></symbol>
    <symbol id="i-filter" viewBox="0 0 24 24"><path d="M4 6h16M7 12h10m-7 6h4"/></symbol>
    <symbol id="i-download" viewBox="0 0 24 24"><path d="M12 4v12m-4-4 4 4 4-4M4 20h16"/></symbol>
  </svg>
</template>
