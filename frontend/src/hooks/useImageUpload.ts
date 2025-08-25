import { useState, useCallback } from 'react';

interface UseImageUploadOptions {
  onSuccess?: (imageUrl: string) => void;
  onError?: (error: string) => void;
}

interface UseImageUploadReturn {
  selectedFile: File | null;
  isUploading: boolean;
  error: string | null;
  preview: string | null;
  handleFileSelect: (file: File | null) => void;
  uploadImage: (agentId: string) => Promise<void>;
  reset: () => void;
}

export function useImageUpload(options: UseImageUploadOptions = {}): UseImageUploadReturn {
  const [selectedFile, setSelectedFile] = useState<File | null>(null);
  const [isUploading, setIsUploading] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [preview, setPreview] = useState<string | null>(null);

  const handleFileSelect = useCallback((file: File | null) => {
    setSelectedFile(file);
    setError(null);
    
    if (file) {
      // Create preview
      const reader = new FileReader();
      reader.onload = (e) => {
        setPreview(e.target?.result as string);
      };
      reader.readAsDataURL(file);
    } else {
      setPreview(null);
    }
  }, []);

  const uploadImage = useCallback(async (agentId: string) => {
    if (!selectedFile) {
      setError('No file selected');
      return;
    }

    setIsUploading(true);
    setError(null);

    try {
      const formData = new FormData();
      formData.append('file', selectedFile);

      const response = await fetch(`/api/agents/${agentId}/image`, {
        method: 'POST',
        body: formData,
      });

      if (!response.ok) {
        const errorText = await response.text();
        throw new Error(errorText || 'Upload failed');
      }

      const result = await response.json();
      const imageUrl = `/api/agents/${agentId}/image`;
      
      options.onSuccess?.(imageUrl);
      
      // Reset state after successful upload
      setSelectedFile(null);
      setPreview(null);
    } catch (err) {
      const errorMessage = err instanceof Error ? err.message : 'Upload failed';
      setError(errorMessage);
      options.onError?.(errorMessage);
    } finally {
      setIsUploading(false);
    }
  }, [selectedFile, options]);

  const reset = useCallback(() => {
    setSelectedFile(null);
    setIsUploading(false);
    setError(null);
    setPreview(null);
  }, []);

  return {
    selectedFile,
    isUploading,
    error,
    preview,
    handleFileSelect,
    uploadImage,
    reset,
  };
}