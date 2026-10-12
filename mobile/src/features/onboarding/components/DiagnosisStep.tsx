import React, { useState } from 'react';
import { View, Text, Pressable } from 'react-native';
import { observer } from '@legendapp/state/react';
import * as Haptics from 'expo-haptics';
import { onboardingState$ } from '../onboardingState';

const OPTIONS = [
  {
    id: 'A',
    label: 'My mind was completely loud and chaotic.',
    subtitle: 'High ambient noise, restless thoughts',
  },
  {
    id: 'B',
    label: 'I felt a sudden sense of relief.',
    subtitle: 'Needed space to breathe and disconnect',
  },
  {
    id: 'C',
    label: 'It was hard to stay still for just 10 seconds.',
    subtitle: 'Conditioned dopamine loop urgency',
  },
];

export const DiagnosisStep: React.FC = observer(() => {
  const selectedDiagnosis = onboardingState$.userDiagnosis.get();
  const [selectedId, setSelectedId] = useState<string | null>(null);

  const handleSelect = (optionText: string, optionId: string) => {
    setSelectedId(optionId);
    Haptics.impactAsync(Haptics.ImpactFeedbackStyle.Medium);
    onboardingState$.userDiagnosis.set(optionText);

    // 400ms delayed transition to Step 3
    setTimeout(() => {
      onboardingState$.currentStep.set(3);
    }, 400);
  };

  return (
    <View className="flex-1 bg-[#0A0A0A] justify-between px-6 pt-12 pb-10">
      <View className="mt-5 items-start">
        <Text className="text-[10px] font-extrabold text-[#3EF090] tracking-[2px] mb-3 uppercase">
          STEP 02 // DIAGNOSIS
        </Text>
        <Text className="text-3xl font-black text-white tracking-tight mb-2">
          How did that feel?
        </Text>
        <Text className="text-sm leading-5 text-[#999999]">
          Select the statement that best describes your internal state during those 10 seconds.
        </Text>
      </View>

      <View className="gap-4 my-5">
        {OPTIONS.map((opt) => {
          const isSelected = selectedDiagnosis === opt.label || selectedId === opt.id;
          return (
            <Pressable
              key={opt.id}
              className={`p-5 rounded-2xl border transition-all ${
                isSelected
                  ? 'bg-[#16281E] border-[#3EF090] shadow-md shadow-[#3EF090]/20'
                  : 'bg-[#121212] border-[#222222]'
              }`}
              onPress={() => handleSelect(opt.label, opt.id)}
            >
              <View className="flex-row items-start gap-3.5">
                <View
                  className={`w-8 h-8 rounded-xl items-center justify-center mt-0.5 ${
                    isSelected ? 'bg-[#3EF090]' : 'bg-[#222222]'
                  }`}
                >
                  <Text
                    className={`text-sm font-extrabold ${
                      isSelected ? 'text-[#0A0A0A]' : 'text-[#888888]'
                    }`}
                  >
                    {opt.id}
                  </Text>
                </View>
                <View className="flex-1">
                  <Text
                    className={`text-base font-bold leading-5 mb-1 ${
                      isSelected ? 'text-white' : 'text-[#E0E0E0]'
                    }`}
                  >
                    "{opt.label}"
                  </Text>
                  <Text className="text-xs text-[#666666]">{opt.subtitle}</Text>
                </View>
              </View>
            </Pressable>
          );
        })}
      </View>

      <View className="items-center">
        <Text className="text-[11px] text-[#444444] text-center font-medium">
          Your baseline diagnosis calibrates your Suach experience.
        </Text>
      </View>
    </View>
  );
});
