# mygin 061701
server package  base on gin

# 入参签名 与 服务端检测的方案
## Header 字段

Next.js 调用 Go API 时，需要额外增加以下 header：

```text
X-App-Id: admin-web
X-Timestamp: 1720000000000
X-Nonce: 7e35f2f0b8c94a3a
X-Body-SHA256: body hash
X-Signature: hmac signature
```

字段说明：

| Header | 说明 |
| --- | --- |
| `X-App-Id` | 调用方身份，例如 `admin-web`、`client-web` |
| `X-Timestamp` | 毫秒时间戳 |
| `X-Nonce` | 每次请求生成的随机字符串 |
| `X-Body-SHA256` | 原始请求 body 的 SHA256 hex |
| `X-Signature` | 使用密钥生成的 HMAC-SHA256 hex |

## 签名字符串

签名字符串固定为 7 行：

```text
HTTP_METHOD
REQUEST_PATH
CANONICAL_QUERY
BODY_SHA256
TIMESTAMP
NONCE
APP_ID
```

示例：

```text
POST
/admin/api/order/push

8f3d2d...
1720000000000
7e35f2f0b8c94a3a
admin-web
```

第三行为空，表示该请求没有 query 参数。

最终签名：

```text
X-Signature = HMAC-SHA256(secret, canonicalString)
```

## Query 标准化规则

GET 请求的参数通常在 URL query 中，所以 query 必须参与签名。

标准化规则：

1. 解析 `RawQuery`
2. key 按字典序排序
3. 同一个 key 有多个 value 时，value 也排序
4. 空值保留
5. 使用 URL query encode
6. 拼接为 `key=value&key2=value2`

示例：

原始 URL：

```text
/admin/api/order?page=1&limit=20&business_id=46
```

标准化后：

```text
business_id=46&limit=20&page=1
```

待签名字符串：

```text
GET
/admin/api/order
business_id=46&limit=20&page=1
e3b0c44298fc1c149afbf4c8996fb924...
1720000000000
7e35f2f0b8c94a3a
admin-web
```

其中 `e3b0...` 是空 body 的 SHA256。

## Body Hash 规则

`X-Body-SHA256` 必须基于原始 body bytes 计算，不要基于结构体映射后的对象计算。

例如以下两个 body 在业务结构体里可能都得到默认值：

```json
{}
```

```json
{"order_id":0}
```

但它们的原始 body 不同，所以 hash 也不同。签名层只关心 HTTP 原始请求是否被篡改；业务参数是否合法，由 Go API 的结构体绑定和 validate 处理。

GET 没有 body 时：

```text
body = ""
X-Body-SHA256 = sha256("")
```

## Path 规则

签名中的 `REQUEST_PATH` 必须和 Go API 服务端实际收到的 path 一致。

例如 Next.js 请求：

```text
POST https://api.bcoin-api.com/admin/api/order/push
```

参与签名的 path 应为：

```text
/admin/api/order/push
```

如果 nginx 对路径做了 rewrite，可能导致 Next.js 签名时的 path 和 Go API 收到的 path 不一致，从而验签失败。上线前需要确认 Go 服务实际收到的 `request.URL.Path`。

## TypeScript 示例

以下示例适合放在 Next.js 服务端工具文件中，不要被 `"use client"` 文件 import。

```ts
import crypto from "node:crypto";

const SIGN_HEADERS = {
  appId: "X-App-Id",
  timestamp: "X-Timestamp",
  nonce: "X-Nonce",
  bodySHA256: "X-Body-SHA256",
  signature: "X-Signature",
};

function sha256Hex(data: string | Buffer): string {
  return crypto.createHash("sha256").update(data).digest("hex");
}

function hmacSHA256Hex(secret: string, data: string): string {
  return crypto.createHmac("sha256", secret).update(data).digest("hex");
}

function nonce(): string {
  return crypto.randomBytes(16).toString("hex");
}

function canonicalQuery(rawQuery: string): string {
  if (!rawQuery) return "";

  const params = new URLSearchParams(rawQuery);
  const entries: Array<[string, string]> = [];

  for (const [key, value] of params.entries()) {
    entries.push([key, value]);
  }

  entries.sort(([ak, av], [bk, bv]) => {
    if (ak === bk) return av.localeCompare(bv);
    return ak.localeCompare(bk);
  });

  return entries
    .map(([key, value]) => `${encodeURIComponent(key)}=${encodeURIComponent(value)}`)
    .join("&");
}

export function buildSignHeaders(input: {
  secret: string;
  appId: string;
  method: string;
  path: string;
  rawQuery?: string;
  body?: string | Buffer;
}): Record<string, string> {
  const timestamp = Date.now().toString();
  const requestNonce = nonce();
  const body = input.body ?? "";
  const bodyHash = sha256Hex(body);
  const query = canonicalQuery(input.rawQuery ?? "");

  const canonicalString = [
    input.method.toUpperCase(),
    input.path,
    query,
    bodyHash,
    timestamp,
    requestNonce,
    input.appId,
  ].join("\n");

  return {
    [SIGN_HEADERS.appId]: input.appId,
    [SIGN_HEADERS.timestamp]: timestamp,
    [SIGN_HEADERS.nonce]: requestNonce,
    [SIGN_HEADERS.bodySHA256]: bodyHash,
    [SIGN_HEADERS.signature]: hmacSHA256Hex(input.secret, canonicalString),
  };
}
```

## POST 调用示例

```ts
const body = new URLSearchParams({ order_id: "16145" }).toString();
const path = "/admin/api/order/push";

const signHeaders = buildSignHeaders({
  secret: process.env.ADMIN_API_SIGN_SECRET!,
  appId: "admin-web",
  method: "POST",
  path,
  body,
});

await fetch(`${process.env.ADMIN_API_URL}${path}`, {
  method: "POST",
  headers: {
    ...signHeaders,
    "Content-Type": "application/x-www-form-urlencoded",
  },
  body,
});
```

注意：签名时使用的 `body` 和 `fetch` 发送的 `body` 必须完全一致。

## GET 调用示例

```ts
const path = "/admin/api/order";
const query = new URLSearchParams({
  page: "1",
  limit: "20",
  business_id: "46",
}).toString();

const signHeaders = buildSignHeaders({
  secret: process.env.ADMIN_API_SIGN_SECRET!,
  appId: "admin-web",
  method: "GET",
  path,
  rawQuery: query,
});

await fetch(`${process.env.ADMIN_API_URL}${path}?${query}`, {
  method: "GET",
  headers: signHeaders,
});
```

注意：签名时使用的 `rawQuery` 和最终 URL 中的 query 必须表达相同参数。为了减少差异，建议用同一个 `URLSearchParams` 生成 query。

## 环境变量

服务端环境变量示例：

```env
ADMIN_API_SIGN_APP_ID=admin-web
ADMIN_API_SIGN_SECRET=replace-with-random-secret
CLIENT_API_SIGN_APP_ID=client-web
CLIENT_API_SIGN_SECRET=replace-with-random-secret
```

不要使用 `NEXT_PUBLIC_` 前缀保存密钥。凡是 `NEXT_PUBLIC_` 开头的环境变量都可能被打包到浏览器端。

错误示例：

```env
NEXT_PUBLIC_ADMIN_API_SIGN_SECRET=xxx
```

正确示例：

```env
ADMIN_API_SIGN_SECRET=xxx
```

## 常见错误

### 1. 签名通过不了

优先检查：

- Next.js 签名用的 path 是否等于 Go API 实际收到的 path
- nginx 是否 rewrite 了 path
- query 是否标准化一致
- 签名时的 body 是否和实际发送 body 完全一致
- `Content-Type` 是否导致 body 生成方式变化
- 服务端和 Next.js 服务器时间是否相差超过 5 分钟

### 2. GET 请求没有 body 是否还要签名

要。GET 的 body hash 使用空字符串 hash，但 query 会参与签名。

### 3. 用户改了参数会怎样

如果用户直接调用 Go API 并修改参数，由于没有服务端 secret，无法生成合法 `X-Signature`，Go API 会拒绝请求。

### 4. 用户重复提交同一个请求会怎样

基础签名可以通过 `timestamp` 限制旧请求。若要严格防重复提交，需要 Go API 结合 Redis 保存 `nonce`，同一个 `app_id + nonce` 在有效期内只能使用一次。

### 5. 浏览器能看到签名密钥吗

不能，前提是签名代码只在 Next.js 服务端执行，并且密钥没有使用 `NEXT_PUBLIC_` 前缀。

## 对接检查清单

- [ ] 签名逻辑只放在 Next.js 服务端代码
- [ ] 密钥只使用服务端环境变量
- [ ] `method`、`path`、`query`、`body` 与实际请求一致
- [ ] GET 请求把 query 纳入签名
- [ ] POST 请求签名 body 和实际发送 body 完全一致
- [ ] nginx 不改写签名 path，或双方明确约定签名 path
- [ ] Go API 中间件在业务参数绑定前完成验签
- [ ] 后续如需防重放，接入 Redis nonce 校验
