import {
  Alert,
  Badge,
  Box,
  Button,
  Group,
  Text,
  Title,
} from "@mantine/core";
import { useDisclosure } from "@mantine/hooks";
import { Hero } from "../../components/Hero";
import { PageWrapper } from "../../components/PageWrapper";
import { useLocation } from "wouter";
import { useExecutablesPage } from "./useExecutablesPage";
import { ExecutablesTable, ViewMode } from "./ExecutablesTable";
import { FilterButton, FilterDrawer } from "./Filters";
import classes from "./Executables.module.css";
import { useState } from "react";
import { IconArticle, IconList, IconRefresh } from "@tabler/icons-react";

export function Executables() {
  const {
    executables,
    isLoading,
    error,
    filterState,
    onFilterChange,
    onClearAll,
    activeFilterCount,
    workspaceOptions,
    namespaceOptions,
    tagOptions,
    verbOptions,
    refresh,
  } = useExecutablesPage();

  if (error) {
    console.log(error);
  }

  const [opened, { open, close }] = useDisclosure(false);
  const [viewMode, setViewMode] = useState<ViewMode>(ViewMode.Card);
  const [, setLocation] = useLocation();

  const handleRowClick = (ref: string) =>
    setLocation(`/executable/${encodeURIComponent(ref)}`);

  return (
    <PageWrapper>
      <Hero variant="split" pattern="subtle">
        <Hero.Header>
          <Title order={2}>Executables</Title>
          <Text c="dimmed">
            Discover and run your development workflows
          </Text>
        </Hero.Header>
        <Hero.Actions>
          <Badge size="sm" c="dimmed" color="tertiary" variant="dot">
            {isLoading ? "Loading..." : `${executables.length} total`}
          </Badge>
          <Button size="compact-xs" color="tertiary" leftSection={<IconRefresh size={16} />} onClick={refresh}>
            Scan Workspaces
          </Button>
        </Hero.Actions>
      </Hero>

      <Group mt="md" gap="sm">
        <Box>
          <Button.Group>
            <Button
              size="xs"
              variant={viewMode === ViewMode.Card ? "filled" : "light"}
              onClick={() => setViewMode(ViewMode.Card)}
            >
              <IconArticle size={16} />
            </Button>
            <Button
              size="xs"
              variant={viewMode === ViewMode.List ? "filled" : "light"}
              onClick={() => setViewMode(ViewMode.List)}
            >
              <IconList size={16} />
            </Button>
          </Button.Group>
        </Box>
        <Box>
          <FilterButton activeCount={activeFilterCount} onClick={open} />
        </Box>
      </Group>

      {error && (
        <Alert title="Error" variant="light" color="red.5">
          Encountered and error while loading executables:
          {error.name}
        </Alert>
      )}

      <Box mt="lg" className={classes.tableContainer}>
        <ExecutablesTable
          items={isLoading ? [] : executables}
          loading={isLoading}
          hasActiveFilters={activeFilterCount > 0}
          onRowClick={handleRowClick}
          viewMode={viewMode}
        />
      </Box>

      <FilterDrawer
        opened={opened}
        onClose={close}
        filterState={filterState}
        onFilterChange={onFilterChange}
        onClearAll={onClearAll}
        workspaceOptions={workspaceOptions}
        namespaceOptions={namespaceOptions}
        tagOptions={tagOptions}
        verbOptions={verbOptions}
      />
    </PageWrapper>
  );
}
