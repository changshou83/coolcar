export const formatTime = (date: Date) => {
  const year = date.getFullYear();
  const month = date.getMonth() + 1;
  const day = date.getDate();
  const hour = date.getHours();
  const minute = date.getMinutes();
  const second = date.getSeconds();

  return (
    [year, month, day].map(formatNumber).join("/") +
    " " +
    [hour, minute, second].map(formatNumber).join(":")
  );
};

export const formatDate = (millis: number) => {
  const dt = new Date(millis);
  const y = dt.getFullYear();
  const m = dt.getMonth() + 1;
  const d = dt.getDate();
  return `${padString(y)}-${padString(m)}-${padString(d)}`;
};

export function formatDuration(sec: number) {
  const h = Math.floor(sec / 3600);
  sec -= 3600 * h;
  const m = Math.floor(sec / 60);
  sec -= 60 * m;
  const s = Math.floor(sec);
  return {
    hh: padString(h),
    mm: padString(m),
    ss: padString(s),
  };
}

export function formatElapsed(sec: number) {
  const dur = formatDuration(sec);
  return `${dur.hh}:${dur.mm}:${dur.ss}`;
}

export function formatFee(cents: number) {
  return (cents / 100).toFixed(2);
}

const formatNumber = (n: number) => {
  return n.toString().padStart(2, "0");
};

function padString(n: number) {
  return n.toFixed(0).padStart(2, "0");
}
