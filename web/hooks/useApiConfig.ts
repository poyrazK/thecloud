'use client';

import { useCallback, useEffect, useMemo, useState } from 'react';
import {
  type CloudApiConfig,
  getStoredCloudApiConfig,
  saveStoredCloudApiConfig,
} from '@/lib/api';

export function useApiConfig() {
  const [config, setConfig] = useState<CloudApiConfig>(() => getStoredCloudApiConfig());
  const ready = typeof window !== 'undefined';

  useEffect(() => {
    const syncFromStorage = () => {
      setConfig(getStoredCloudApiConfig());
    };

    window.addEventListener('storage', syncFromStorage);
    return () => {
      window.removeEventListener('storage', syncFromStorage);
    };
  }, []);

  const updateConfig = useCallback((update: Partial<CloudApiConfig>) => {
    const next = saveStoredCloudApiConfig(update);
    setConfig(next);
    return next;
  }, []);

  const hasCredentials = useMemo(() => Boolean(config.apiKey), [config.apiKey]);

  return {
    config,
    ready,
    hasCredentials,
    updateConfig,
  };
}
