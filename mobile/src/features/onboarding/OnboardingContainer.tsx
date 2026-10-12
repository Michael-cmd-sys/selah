import React from 'react';
import { View, Pressable, Text } from 'react-native';
import { observer } from '@legendapp/state/react';
import { StatusBar } from 'expo-status-bar';
import { Audio } from 'expo-av';
import * as Haptics from 'expo-haptics';

import { onboardingState$ } from './onboardingState';
import { SilenceTestStep } from './components/SilenceTestStep';
import { DiagnosisStep } from './components/DiagnosisStep';
import { TimeAndTribeStep } from './components/TimeAndTribeStep';
import { NarrativeStep } from './components/NarrativeStep';
import { CovenantLockStep } from './components/CovenantLockStep';

const TOTAL_STEPS = 5;

// Module-level audio pre-warm on demand
let audioInitialized = false;
async function ensureAudioMode() {
  if (audioInitialized) return;
  audioInitialized = true;
  try {
    await Audio.setAudioModeAsync({
      playsInSilentModeIOS: true,
      staysActiveInBackground: false,
      shouldDuckAndroid: true,
    });
  } catch (err) {
    console.warn('Audio initialization warning:', err);
  }
}

export const OnboardingContainer: React.FC = observer(() => {
  const currentStep = onboardingState$.currentStep.get();

  // Lazy audio mode setup on render
  ensureAudioMode();

  const handleStepBack = () => {
    if (currentStep > 1) {
      Haptics.impactAsync(Haptics.ImpactFeedbackStyle.Light);
      onboardingState$.currentStep.set(currentStep - 1);
    }
  };

  const renderStepContent = () => {
    switch (currentStep) {
      case 1:
        return <SilenceTestStep />;
      case 2:
        return <DiagnosisStep />;
      case 3:
        return <TimeAndTribeStep />;
      case 4:
        return <NarrativeStep />;
      case 5:
        return <CovenantLockStep />;
      default:
        return <SilenceTestStep />;
    }
  };

  return (
    <SafeAreaView className="flex-1 bg-[#0A0A0A]">
      <StatusBar style="light" />
      <View className="flex-1 bg-[#0A0A0A]">
        {/* Top Header Navigation Bar */}
        <View className="flex-row items-center justify-between px-6 pt-3 pb-2">
          {currentStep > 1 ? (
            <Pressable className="w-14 py-1.5" onPress={handleStepBack}>
              <Text className="text-sm font-bold text-[#3EF090]">‹ Back</Text>
            </Pressable>
          ) : (
            <View className="w-[50px]" />
          )}

          {/* Staggered Step Indicators */}
          <View className="flex-1 flex-row gap-1.5 mx-4">
            {Array.from({ length: TOTAL_STEPS }).map((_, index) => {
              const stepNum = index + 1;
              const isActive = stepNum === currentStep;
              const isPast = stepNum < currentStep;
              return (
                <View
                  key={stepNum}
                  className={`flex-1 h-1 rounded-sm ${isPast
                      ? 'bg-[#1E4D34]'
                      : isActive
                        ? 'bg-[#3EF090]'
                        : 'bg-[#222222]'
                    }`}
                />
              );
            })}
          </View>

          <Text className="text-[11px] font-extrabold text-[#666666] w-7 text-right">
            {currentStep}/{TOTAL_STEPS}
          </Text>
        </View>

        {/* Step Content Wrapper with fixed bounds */}
        <View className="flex-1">{renderStepContent()}</View>
      </View>
    </SafeAreaView>
  );
});
