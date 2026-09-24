import type { NextConfig } from "next";

import { loadPublicEnvironment } from "./lib/environment";

const publicEnvironment = loadPublicEnvironment();

const nextConfig: NextConfig = {
  env: {
    NEXT_PUBLIC_MANORECK_API_BASE_URL: publicEnvironment.apiBaseUrl,
  },
  output: "standalone",
  poweredByHeader: false,
  reactStrictMode: true,
};

export default nextConfig;
