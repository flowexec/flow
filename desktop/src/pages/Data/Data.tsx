import { Tabs } from "@mantine/core";
import { IconBraces, IconLock } from "@tabler/icons-react";
import {PageWrapper} from "../../components/PageWrapper.tsx";

export function Data() {
  return (
      <PageWrapper>
    <Tabs defaultValue="cache">
      <Tabs.List>
        <Tabs.Tab value="cache" leftSection={<IconBraces size={12} />}>
          Cache
        </Tabs.Tab>
        <Tabs.Tab value="vault" leftSection={<IconLock size={12} />}>
          Vault
        </Tabs.Tab>
      </Tabs.List>
      <Tabs.Panel value="cache">Cache data should show here</Tabs.Panel>
      <Tabs.Panel value="vault">Vault data should show here</Tabs.Panel>
    </Tabs>
      </PageWrapper>
  );
}
