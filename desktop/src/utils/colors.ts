// djb2 string hash → 32-bit signed int
function stringHash(str: string) {
    let h = 5381;
    for (let i = 0; i < str.length; i++) h = ((h << 5) + h) ^ str.charCodeAt(i);
    return h >>> 0; // unsigned
}

// HSL → hex helpers
function hslToRgb(h : number, s : number, l : number) {
    s /= 100; l /= 100;
    const k = n => (n + h / 30) % 12;
    const a = s * Math.min(l, 1 - l);
    const f = n => l - a * Math.max(-1, Math.min(k(n) - 3, Math.min(9 - k(n), 1)));
    return [Math.round(255 * f(0)), Math.round(255 * f(8)), Math.round(255 * f(4))];
}
function rgbToHex([r, g, b] : number[]) {
    return '#' + [r, g, b].map(x => x.toString(16).padStart(2, '0')).join('');
}

// Main: string → hex color
function stringToColor(str : string) {
    const h = stringHash(str);
    const hue = h % 360;                  // 0–359
    const sat = 65;                       // tweak for vibrancy
    const light = 55;                     // tweak for readability
    return rgbToHex(hslToRgb(hue, sat, light));
}

// Optional: pick black/white text for contrast
function idealTextColor(bgHex : string) {
    const r = parseInt(bgHex.slice(1,3),16);
    const g = parseInt(bgHex.slice(3,5),16);
    const b = parseInt(bgHex.slice(5,7),16);
    const luminance = 0.2126*r + 0.7152*g + 0.0722*b;
    return luminance > 140 ? '#000000' : '#FFFFFF';
}

// Example
const color = stringToColor('alice@example.com'); // e.g. "#d15f8c"
const text  = idealTextColor(color);
