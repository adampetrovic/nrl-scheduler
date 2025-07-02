import React from 'react';

interface ConstraintWeightSliderProps {
  value: number;
  onChange: (value: number) => void;
  min?: number;
  max?: number;
  step?: number;
  label?: string;
  disabled?: boolean;
}

export const ConstraintWeightSlider: React.FC<ConstraintWeightSliderProps> = ({
  value,
  onChange,
  min = 0,
  max = 1,
  step = 0.1,
  label = 'Weight',
  disabled = false,
}) => {
  const handleChange = (event: React.ChangeEvent<HTMLInputElement>) => {
    onChange(parseFloat(event.target.value));
  };

  return (
    <div className="w-full">
      <div className="flex justify-between items-center mb-2">
        <label className="text-sm font-medium text-gray-700">{label}</label>
        <span className="text-sm text-gray-500 bg-gray-100 px-2 py-1 rounded">
          {value.toFixed(1)}
        </span>
      </div>
      <div className="relative">
        <input
          type="range"
          min={min}
          max={max}
          step={step}
          value={value}
          onChange={handleChange}
          disabled={disabled}
          className={`w-full h-2 bg-gray-200 rounded-lg appearance-none cursor-pointer slider ${
            disabled ? 'opacity-50 cursor-not-allowed' : ''
          }`}
        />
        <div className="flex justify-between text-xs text-gray-500 mt-1">
          <span>Low ({min})</span>
          <span>High ({max})</span>
        </div>
      </div>
      <style dangerouslySetInnerHTML={{
        __html: `
          .slider::-webkit-slider-thumb {
            appearance: none;
            height: 20px;
            width: 20px;
            border-radius: 50%;
            background: #3b82f6;
            cursor: pointer;
            box-shadow: 0 2px 4px rgba(0, 0, 0, 0.2);
          }
          
          .slider::-moz-range-thumb {
            height: 20px;
            width: 20px;
            border-radius: 50%;
            background: #3b82f6;
            cursor: pointer;
            border: none;
            box-shadow: 0 2px 4px rgba(0, 0, 0, 0.2);
          }
          
          .slider:disabled::-webkit-slider-thumb {
            background: #9ca3af;
            cursor: not-allowed;
          }
          
          .slider:disabled::-moz-range-thumb {
            background: #9ca3af;
            cursor: not-allowed;
          }
        `
      }} />
    </div>
  );
};