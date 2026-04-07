'use client';

import { useState } from 'react';
import toast from 'react-hot-toast';
import { motion } from 'framer-motion';
import { apiClient } from '@/lib/api-client';
import type { DataSource } from '@/types';
import DataSourceList from './DataSourceList';
import DataSourceForm from './DataSourceForm';
import Modal from '../Modal';
import { staggerContainer, staggerItem, springConfig } from '@/lib/animations';

export default function DataSourceManager() {
  const [view, setView] = useState<'list' | 'create' | 'edit'>('list');
  const [selectedDataSource, setSelectedDataSource] = useState<DataSource | null>(null);
  const [refreshKey, setRefreshKey] = useState(0);

  const handleCreateNew = () => {
    setSelectedDataSource(null);
    setView('create');
  };

  const handleEditDataSource = (dataSource: DataSource) => {
    setSelectedDataSource(dataSource);
    setView('edit');
  };

  const handleSave = () => {
    setView('list');
    setSelectedDataSource(null);
    setRefreshKey((prev) => prev + 1);
  };

  const handleCancel = () => {
    setView('list');
    setSelectedDataSource(null);
  };

  return (
    <motion.div 
      className="max-w-[1600px] mx-auto space-y-8 pb-12 px-4 md:px-6"
      initial="hidden"
      animate="visible"
      variants={staggerContainer}
    >
      {/* Header */}
      <motion.div 
        className="flex flex-col md:flex-row md:items-center justify-between gap-6 pt-4"
        variants={staggerItem}
      >
        <div className="space-y-1">
          <h1 className="text-4xl font-bold tracking-tight text-slate-900 dark:text-white">
            Data Sources
          </h1>
          <p className="text-slate-500 dark:text-slate-400 font-medium">
            Manage connections to your databases.
          </p>
        </div>
        
        <motion.div 
          className="flex items-center gap-4"
          variants={staggerItem}
        >
          <motion.button
            onClick={handleCreateNew}
            className="btn btn-primary h-11 px-8 rounded-2xl text-sm font-bold sleek-shadow"
            whileHover={{ scale: 1.02 }}
            whileTap={{ scale: 0.98 }}
            transition={springConfig.micro}
          >
            <span className="text-xl mr-2">+</span>
            Add Data Source
          </motion.button>
        </motion.div>
      </motion.div>

      {/* Content */}
      <motion.div 
        className="space-y-6"
        variants={staggerItem}
      >
        <DataSourceList key={refreshKey} onEditDataSource={handleEditDataSource} selectedId={null} />
      </motion.div>

      <Modal 
        isOpen={view === 'create' || view === 'edit'} 
        onClose={handleCancel}
        title={view === 'create' ? 'Add Data Source' : 'Edit Data Source'}
      >
        <DataSourceForm
          dataSource={selectedDataSource || undefined}
          onSave={handleSave}
          onCancel={handleCancel}
        />
      </Modal>
    </motion.div>
  );
}
