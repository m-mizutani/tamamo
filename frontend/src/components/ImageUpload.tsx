import React, { useState, useRef, useCallback } from 'react';
import { X, Image as ImageIcon } from 'lucide-react';
import { Button } from './ui/button';
import { Card } from './ui/card';

interface ImageUploadProps {
  onImageSelect: (file: File | null) => void;
  previewUrl?: string | null;
  isUploading?: boolean;
  error?: string | null;
  maxFileSize?: number; // in MB
  acceptedTypes?: string[];
}

export function ImageUpload({
  onImageSelect,
  previewUrl,
  isUploading = false,
  error,
  maxFileSize = 10,
  acceptedTypes = ['image/jpeg', 'image/png']
}: ImageUploadProps) {
  const [dragActive, setDragActive] = useState(false);
  const [validationError, setValidationError] = useState<string | null>(null);
  const inputRef = useRef<HTMLInputElement>(null);

  const validateFile = (file: File): string | null => {
    // Check file type
    if (!acceptedTypes.includes(file.type)) {
      return `Please select a valid image file (${acceptedTypes.join(', ')})`;
    }

    // Check file size
    const maxSizeBytes = maxFileSize * 1024 * 1024;
    if (file.size > maxSizeBytes) {
      return `File size must be less than ${maxFileSize}MB`;
    }

    return null;
  };

  const handleFile = useCallback((file: File) => {
    const fileValidationError = validateFile(file);
    if (fileValidationError) {
      setValidationError(fileValidationError);
      onImageSelect(null);
      return;
    }

    // Clear validation error if file is valid
    setValidationError(null);
    onImageSelect(file);
  }, [onImageSelect, maxFileSize, acceptedTypes]);

  const handleDrag = useCallback((e: React.DragEvent) => {
    e.preventDefault();
    e.stopPropagation();
    if (e.type === 'dragenter' || e.type === 'dragover') {
      setDragActive(true);
    } else if (e.type === 'dragleave') {
      setDragActive(false);
    }
  }, []);

  const handleDrop = useCallback((e: React.DragEvent) => {
    e.preventDefault();
    e.stopPropagation();
    setDragActive(false);

    if (e.dataTransfer.files && e.dataTransfer.files[0]) {
      handleFile(e.dataTransfer.files[0]);
    }
  }, [handleFile]);

  const handleChange = useCallback((e: React.ChangeEvent<HTMLInputElement>) => {
    e.preventDefault();
    if (e.target.files && e.target.files[0]) {
      handleFile(e.target.files[0]);
    }
  }, [handleFile]);

  const removeImage = useCallback(() => {
    setValidationError(null);
    onImageSelect(null);
    if (inputRef.current) {
      inputRef.current.value = '';
    }
  }, [onImageSelect]);

  const openFileDialog = useCallback(() => {
    inputRef.current?.click();
  }, []);

  return (
    <div className="w-full">
      <Card className="p-4">
        <div className="space-y-4">
          <h3 className="text-sm font-medium">Agent Image</h3>
          
          {previewUrl ? (
            <div className="relative">
              <div className="relative w-32 h-32 mx-auto rounded-lg overflow-hidden border-2 border-gray-200">
                <img
                  src={previewUrl}
                  alt="Agent preview"
                  className="w-full h-full object-cover"
                />
                {!isUploading && (
                  <button
                    onClick={removeImage}
                    className="absolute top-1 right-1 p-1 bg-red-500 text-white rounded-full hover:bg-red-600 transition-colors"
                    disabled={isUploading}
                  >
                    <X size={12} />
                  </button>
                )}
                {isUploading && (
                  <div className="absolute inset-0 bg-black bg-opacity-50 flex items-center justify-center">
                    <div className="animate-spin rounded-full h-6 w-6 border-b-2 border-white"></div>
                  </div>
                )}
              </div>
              <Button
                type="button"
                variant="outline"
                size="sm"
                onClick={openFileDialog}
                disabled={isUploading}
                className="w-full mt-2"
              >
                Change Image
              </Button>
            </div>
          ) : (
            <div
              className={`
                relative border-2 border-dashed rounded-lg p-6 text-center cursor-pointer transition-colors
                ${dragActive ? 'border-blue-400 bg-blue-50' : 'border-gray-300'}
                ${isUploading ? 'opacity-50 cursor-not-allowed' : 'hover:border-gray-400'}
              `}
              onDragEnter={handleDrag}
              onDragLeave={handleDrag}
              onDragOver={handleDrag}
              onDrop={handleDrop}
              onClick={openFileDialog}
            >
              <div className="flex flex-col items-center space-y-2">
                {isUploading ? (
                  <div className="animate-spin rounded-full h-8 w-8 border-b-2 border-gray-600"></div>
                ) : (
                  <ImageIcon className="h-8 w-8 text-gray-400" />
                )}
                <div className="text-sm text-gray-600">
                  {isUploading ? (
                    'Uploading...'
                  ) : (
                    <>
                      <span className="font-medium">Click to upload</span> or drag and drop
                    </>
                  )}
                </div>
                <div className="text-xs text-gray-500">
                  PNG or JPEG up to {maxFileSize}MB
                </div>
              </div>
            </div>
          )}

          <input
            ref={inputRef}
            type="file"
            className="hidden"
            accept={acceptedTypes.join(',')}
            onChange={handleChange}
            disabled={isUploading}
          />

          {(error || validationError) && (
            <div className="text-sm text-red-600 bg-red-50 border border-red-200 rounded-md p-2">
              {error || validationError}
            </div>
          )}
        </div>
      </Card>
    </div>
  );
}