// stats
type statistics = {
  dbsize: number;
  results: number;
  headers: number;
  consolelogs: number;
  networklogs: number;
  response_code_stats: response_code_stats[];
};

interface response_code_stats {
  code: number;
  count: number;
}

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
  probed_at: string;
  title: string;
  response_code: number;
  file_name: string;
  screenshot: string;
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
  san_list: sanlist[];
  issuer: string;
  valid_from: string;
  valid_to: string;
  server_signature_algorithm: number;
  encrypted_client_hello: boolean;
}

interface sanlist {
  id: number;
  tls_id: number;
  value: string;
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
  content: string;
}

interface consolelog {
  id: number;
  resultid: number;
  type: string;
  value: string;
}

interface cookie {
  id: number;
  result_id: number;
  name: string;
  value: string;
  domain: string;
  path: string;
  expires: string; // actually a timestamp
  size: number;
  http_only: boolean;
  secure: boolean;
  session: boolean;
  priority: string;
  source_scheme: string;
  source_port: number;
}

interface detail {
  id: number;
  url: string;
  probed_at: string;
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
  screenshot: string;
  tls: tls;
  technologies: technology[];
  headers: header[];
  network: networklog[];
  console: consolelog[];
  cookies: cookie[];
}

interface searchresult {
  id: number;
  url: string;
  final_url: string;
  response_code: number;
  content_length: number;
  title: string;
  matched_fields: string[];
  file_name: string;
  screenshot: string;
}

interface technologylist {
  technologies: string[];
}

export type {
  statistics,
  wappalyzer,
  gallery,
  list,
  galleryResult,
  tls,
  sanlist,
  technology,
  header,
  networklog,
  consolelog,
  cookie,
  detail,
  searchresult,
  technologylist,
};