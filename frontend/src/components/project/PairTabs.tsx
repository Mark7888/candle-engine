import type { SourcePair } from '../../types/api';

interface PairTabsProps {
  pairs: SourcePair[];
  activePair: SourcePair | null;
  onSelect: (pair: SourcePair) => void;
}

export default function PairTabs({ pairs, activePair, onSelect }: PairTabsProps) {
  if (pairs.length === 0) return null;

  return (
    <div className="flex flex-col space-y-1 min-w-[140px]">
      {pairs.map((pair) => {
        const isActive =
          activePair?.pairId === pair.pairId &&
          activePair?.sourceName === pair.sourceName;
        return (
          <button
            key={`${pair.sourceName}-${pair.pairId}`}
            onClick={() => onSelect(pair)}
            className={`px-3 py-2 text-left rounded-lg transition-colors text-sm ${
              isActive
                ? 'bg-indigo-900/60 text-indigo-300 border border-indigo-600'
                : 'bg-transparent text-gray-400 hover:bg-gray-700 hover:text-gray-200'
            }`}
          >
            <span className="font-medium">{pair.pairId}</span>
            <br />
            <span className="text-xs text-gray-500">{pair.sourceName}</span>
          </button>
        );
      })}
    </div>
  );
}
