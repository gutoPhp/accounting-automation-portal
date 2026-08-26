import { useEffect, useState } from "react";

function currentRoute() {
  return window.location.hash.slice(1) || "home";
}

export function useHashRoute() {
  const [route, setRoute] = useState(currentRoute);

  useEffect(() => {
    const handleRouteChange = () => setRoute(currentRoute());
    window.addEventListener("hashchange", handleRouteChange);

    return () => window.removeEventListener("hashchange", handleRouteChange);
  }, []);

  return route;
}
