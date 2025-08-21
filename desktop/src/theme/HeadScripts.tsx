import { ColorSchemeScript } from "@mantine/core";
import { createPortal } from "react-dom";
import { useEffect, useState } from "react";

export function HeadScripts() {
  const [head, setHead] = useState<HTMLElement | null>(null);

  useEffect(() => {
    setHead(document.head);
  }, []);

  if (!head) return null;

  return createPortal(
    <>
      <ColorSchemeScript />
    </>,
    head
  );
}

