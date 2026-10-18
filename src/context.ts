import { createContext } from 'react';
import type { Dataset } from './data.ts';
export const DatasetContext = createContext<Dataset>('source');
