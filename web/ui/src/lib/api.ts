// stats
type statistics = {
  dbsize: number;
  results: number;
  headers: number;
  consolelogs: number;
  networklogs: number;
};

// wappalyzer
type wappalyzer = {
  [name: string]: string;
};

// gallery
type gallery = {
  results: galleryResult[];
  page: number;
  limit: number;
  total_count: number;
};

type galleryResult = {
  id: number;
  url: string;
  title: string;
  response_code: number;
  file_name: string;
  failed: boolean;
  technologies: string[];
};

// list
type list = {
  id: number;
  url: string;
  final_url: string;
  response_code: number;
  response_reason: string;
  protocol: string;
  content_length: number;
  title: string;
  failed: boolean;
  failed_reason: string;
};

// details
interface tls {
  id: number;
  result_id: number;
  protocol: string;
  key_exchange: string;
  cipher: string;
  subject_name: string;
  san_list: string | null;
  issuer: string;
  valid_from: number;
  valid_to: number;
  server_signature_algorithm: number;
  encrypted_client_hello: boolean;
}

interface technology {
  id: number;
  result_id: number;
  value: string;
}

interface header {
  id: number;
  result_id: number;
  key: string;
  value: string | null;
}

interface networklog {
  id: number;
  result_id: number;
  request_type: number;
  status_code: number;
  url: string;
  remote_ip: string;
  mime_type: string;
  time: string;
  error: string;
}

interface consolelog {
  id: number;
  resultid: number;
  type: string;
  value: string;
}

interface detail {
  id: number;
  url: string;
  final_url: string;
  response_code: number;
  response_reason: string;
  protocol: string;
  content_length: number;
  html: string;
  title: string;
  perception_hash: string;
  file_name: string;
  is_pdf: boolean;
  failed: boolean;
  failed_reason: string;
  tls: tls;
  technologies: technology[];
  headers: header[];
  network: networklog[];
  console: consolelog[];
}

const endpoints = {
  // api base path
  base: {
    path: import.meta.env.VITE_GOWITNESS_API_BASE_URL
      ? import.meta.env.VITE_GOWITNESS_API_BASE_URL
      : `http://localhost:8081/api`,
    returnas: [] // n/a
  },
  // screenshot path
  screenshot: {
    path: `http://localhost:8081/screenshots/`,
    returnas: [] // n/a
  },

  statistics: {
    path: `/statistics`,
    returnas: {} as statistics
  },
  wappalyzer: {
    path: `/wappalyzer`,
    returnas: {} as wappalyzer
  },
  gallery: {
    path: `/gallery`,
    returnas: {} as gallery
  },
  list: {
    path: `/list`,
    returnas: [] as list[]
  },
  detail: {
    path: `/detail/:id`,
    returnas: {} as detail
  }
};

type Endpoints = typeof endpoints;
type EndpointReturnType<K extends keyof Endpoints> = Endpoints[K]['returnas'];

const replacePathParams = (path: string, params?: Record<string, string | number | boolean>): [string, Record<string, string | number | boolean>] => {
  if (!params) return [path, {}]; // If no params provided, return the path as is and an empty object.

  const paramRegex = /:([a-zA-Z0-9_]+)/g;
  const missingParams: string[] = [];
  const remainingParams = { ...params }; // Create a copy of the params object to modify

  // Replace all `:param` placeholders with the corresponding values from params
  const newPath = path.replace(paramRegex, (match, paramName) => {
    if (paramName in remainingParams) {
      const value = remainingParams[paramName];
      delete remainingParams[paramName];
      return encodeURIComponent(value.toString());
    } else {
      missingParams.push(paramName);
      return match;
    }
  });

  // If any required params were missing, throw an error
  if (missingParams.length > 0) {
    throw new Error(`Missing required parameters: ${missingParams.join(', ')}`);
  }

  return [newPath, remainingParams]; // Return the new path and remaining params
};

const serializeParams = (params: Record<string, string | number | boolean>) => {
  const query = new URLSearchParams();
  Object.entries(params).forEach(([key, value]) => {
    query.append(key, value.toString());
  });
  return query.toString() ? `?${query.toString()}` : '';
};

const get = async <K extends keyof Endpoints>(
  endpointKey: K,
  params?: Record<string, string | number | boolean>,
  raw: boolean = false
): Promise<EndpointReturnType<K>> => {

  const endpoint = endpoints[endpointKey];
  const [pathWithParams, remainingParams] = replacePathParams(endpoint.path, params);
  const queryString = remainingParams ? serializeParams(remainingParams) : '';

  const res = await fetch(`${endpoints.base.path}${pathWithParams}${queryString}`);

  if (!res.ok) throw new Error(`HTTP Error: Status: ${res.status}`);

  if (raw) return await res.text() as unknown as EndpointReturnType<K>;
  return await res.json() as EndpointReturnType<K>;
};

export type {
  statistics,
  wappalyzer,
  gallery,
  list,
  galleryResult,
  tls,
  technology,
  header,
  networklog,
  consolelog,
  detail as result,
};
export {
  endpoints,
  get,
};