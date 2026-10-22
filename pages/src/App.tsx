import { Site } from "./Site.tsx";
import { useSite } from "./useSite.ts";

export function App() {
  return <Site {...useSite()} />;
}
