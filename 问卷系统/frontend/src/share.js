// H5 分享链接签发工具（防盗刷）
// ⚠️ SHARE_SECRET 必须与 public/h5/index.html 中的 SHARE_SECRET 完全一致，
//    否则 H5 页面会判定签名无效而锁定。
export const SHARE_SECRET = 'LILIANG_SURVEY_SECRET_2026_CHANGE_ME'

// ---------- 纯 JS HMAC-SHA256（与 h5/index.html 保持一致） ----------
function utf8Bytes(str) {
  const out = []
  for (let i = 0; i < str.length; i++) {
    let c = str.charCodeAt(i)
    if (c < 0x80) out.push(c)
    else if (c < 0x800) out.push(0xc0 | (c >> 6), 0x80 | (c & 0x3f))
    else if (c >= 0xd800 && c <= 0xdbff) {
      const c2 = str.charCodeAt(++i)
      const cp = 0x10000 + ((c & 0x3ff) << 10) + (c2 & 0x3ff)
      out.push(0xf0 | (cp >> 18), 0x80 | ((cp >> 12) & 0x3f), 0x80 | ((cp >> 6) & 0x3f), 0x80 | (cp & 0x3f))
    } else out.push(0xe0 | (c >> 12), 0x80 | ((c >> 6) & 0x3f), 0x80 | (c & 0x3f))
  }
  return out
}
function sha256bytes(msg) {
  const rrot = (x, n) => (x >>> n) | (x << (32 - n))
  const K = [0x428a2f98,0x71374491,0xb5c0fbcf,0xe9b5dba5,0x3956c25b,0x59f111f1,0x923f82a4,0xab1c5ed5,0xd807aa98,0x12835b01,0x243185be,0x550c7dc3,0x72be5d74,0x80deb1fe,0x9bdc06a7,0xc19bf174,0xe49b69c1,0xefbe4786,0x0fc19dc6,0x240ca1cc,0x2de92c6f,0x4a7484aa,0x5cb0a9dc,0x76f988da,0x983e5152,0xa831c66d,0xb00327c8,0xbf597fc7,0xc6e00bf3,0xd5a79147,0x06ca6351,0x14292967,0x27b70a85,0x2e1b2138,0x4d2c6dfc,0x53380d13,0x650a7354,0x766a0abb,0x81c2c92e,0x92722c85,0xa2bfe8a1,0xa81a664b,0xc24b8b70,0xc76c51a3,0xd192e819,0xd6990624,0xf40e3585,0x106aa070,0x19a4c116,0x1e376c08,0x2748774c,0x34b0bcb5,0x391c0cb3,0x4ed8aa4a,0x5b9cca4f,0x682e6ff3,0x748f82ee,0x78a5636f,0x84c87814,0x8cc70208,0x90befffa,0xa4506ceb,0xbef9a3f7,0xc67178f2]
  let h = [0x6a09e667,0xbb67ae85,0x3c6ef372,0xa54ff53a,0x510e527f,0x9b05688c,0x1f83d9ab,0x5be0cd19]
  const l = msg.length
  const pad = msg.concat([0x80])
  while (pad.length % 64 !== 56) pad.push(0)
  const bitLen = l * 8
  pad.push(0, 0, 0, 0, (bitLen >>> 24) & 255, (bitLen >>> 16) & 255, (bitLen >>> 8) & 255, bitLen & 255)
  const w = new Array(64)
  for (let off = 0; off < pad.length; off += 64) {
    for (let j = 0; j < 16; j++) w[j] = (pad[off + j * 4] << 24) | (pad[off + j * 4 + 1] << 16) | (pad[off + j * 4 + 2] << 8) | pad[off + j * 4 + 3]
    for (let j = 16; j < 64; j++) {
      const s0 = rrot(w[j - 15], 7) ^ rrot(w[j - 15], 18) ^ (w[j - 15] >>> 3)
      const s1 = rrot(w[j - 2], 17) ^ rrot(w[j - 2], 19) ^ (w[j - 2] >>> 10)
      w[j] = (w[j - 16] + s0 + w[j - 7] + s1) | 0
    }
    let a = h[0], b = h[1], c = h[2], d = h[3], e = h[4], f = h[5], g = h[6], hh = h[7]
    for (let j = 0; j < 64; j++) {
      const S1 = rrot(e, 6) ^ rrot(e, 11) ^ rrot(e, 25)
      const ch = (e & f) ^ ((~e) & g)
      const t1 = (hh + S1 + ch + K[j] + w[j]) | 0
      const S0 = rrot(a, 2) ^ rrot(a, 13) ^ rrot(a, 22)
      const maj = (a & b) ^ (a & c) ^ (b & c)
      const t2 = (S0 + maj) | 0
      hh = g; g = f; f = e; e = (d + t1) | 0; d = c; c = b; b = a; a = (t1 + t2) | 0
    }
    h[0] = (h[0] + a) | 0; h[1] = (h[1] + b) | 0; h[2] = (h[2] + c) | 0; h[3] = (h[3] + d) | 0
    h[4] = (h[4] + e) | 0; h[5] = (h[5] + f) | 0; h[6] = (h[6] + g) | 0; h[7] = (h[7] + hh) | 0
  }
  return h
}
function hmacSHA256(keyStr, msgStr) {
  let key = utf8Bytes(keyStr)
  if (key.length > 64) {
    const hh = sha256bytes(key)
    key = []
    for (const word of hh) key.push((word >>> 24) & 255, (word >>> 16) & 255, (word >>> 8) & 255, word & 255)
  }
  while (key.length < 64) key.push(0)
  const msg = utf8Bytes(msgStr)
  const ipad = key.map(b => b ^ 0x36)
  const opad = key.map(b => b ^ 0x5c)
  const inner = sha256bytes(ipad.concat(msg))
  const innerBytes = []
  for (const word of inner) innerBytes.push((word >>> 24) & 255, (word >>> 16) & 255, (word >>> 8) & 255, word & 255)
  const outer = sha256bytes(opad.concat(innerBytes))
  return outer.map(x => (x >>> 0).toString(16).padStart(8, '0')).join('')
}

/**
 * 生成带签名与有效期的 H5 分享链接。
 * 签名算法必须与 h5/index.html 的 verifyLink 及后端 /api/h5/projects 一致：
 *   payload = exp + '|' + to + '|' + ids + '|' + pathname
 *   sig     = HMAC-SHA256(SHARE_SECRET, payload)
 * @param {string} h5Url  H5 页面完整地址，如 http://192.168.1.5:5173/h5/index.html
 * @param {number} hours  有效期（小时）；<=0 表示长期有效（exp=0，永不过期）
 * @param {string} to     附加标识（如项目编号列表），会参与签名
 * @param {string} ids    项目 ID 列表（逗号分隔），参与签名；为空表示全部项目
 * @returns {{url:string, exp:number, sig:string, forever:boolean}}
 */
export function buildShareLink(h5Url, hours, to = '', ids = '') {
  const url = new URL(h5Url)
  const h = Number(hours)
  const forever = !isFinite(h) || h <= 0 // 长期有效：exp=0
  const exp = forever ? 0 : Date.now() + Math.round(h * 3600 * 1000)
  const payload = exp + '|' + to + '|' + ids + '|' + url.pathname
  const sig = hmacSHA256(SHARE_SECRET, payload)
  const u = url.origin + url.pathname + '?exp=' + exp + '&sig=' + sig +
    (to ? '&to=' + encodeURIComponent(to) : '') +
    (ids ? '&ids=' + ids : '')
  return { url: u, exp, sig, forever }
}

/** 提取所选项目的 ID 列表（逗号分隔），用于链接签名与后端查询 */
export function encodeIds(projectList) {
  return (projectList || []).map(p => p.id).join(',')
}

/** 复制文本到剪贴板（含降级方案） */
export async function copyText(text) {
  try {
    await navigator.clipboard.writeText(text)
    return true
  } catch (e) {
    const ta = document.createElement('textarea')
    ta.value = text
    ta.style.position = 'fixed'
    ta.style.left = '-9999px'
    document.body.appendChild(ta)
    ta.select()
    let ok = false
    try { ok = document.execCommand('copy') } catch (err) { ok = false }
    document.body.removeChild(ta)
    return ok
  }
}
