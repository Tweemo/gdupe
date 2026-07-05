import type { NextConfig } from "next";

const nextConfig: NextConfig = {
  // Build to static HTML/CSS/JS in `out/`. The app is purely client-side
  // (no SSR/data fetching), so the Go server can serve these files directly
  // alongside its /api routes — one origin, one container.
  output: "export",
};

export default nextConfig;
