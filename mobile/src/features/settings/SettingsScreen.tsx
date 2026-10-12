import React from 'react';
import { View, Text, Pressable, ScrollView, Alert } from 'react-native';
import { observer } from '@legendapp/state/react';
import { useRouter } from 'expo-router';
import AsyncStorage from '@react-native-async-storage/async-storage';
import * as Haptics from 'expo-haptics';
import { onboardingState$, TRIBAL_CLASSES } from '../onboarding/onboardingState';

export const SettingsScreen = observer(function SettingsScreen() {
  const router = useRouter();
  const tribalClass = onboardingState$.tribalClass.get();
  const consistencyTarget = onboardingState$.consistencyTarget.get();
  const morningTime = onboardingState$.morningTime.get();
  const eveningTime = onboardingState$.eveningTime.get();
  const userDiagnosis = onboardingState$.userDiagnosis.get();

  const currentTribalInfo = TRIBAL_CLASSES[consistencyTarget] || TRIBAL_CLASSES[70];

  const handleResetOnboarding = () => {
    Haptics.impactAsync(Haptics.ImpactFeedbackStyle.Medium);
    Alert.alert(
      'Re-Calibrate Onboarding',
      'Are you sure you want to restart your onboarding setup and re-calibrate your tribal identity?',
      [
        { text: 'Cancel', style: 'cancel' },
        {
          text: 'Reset & Re-Calibrate',
          style: 'destructive',
          onPress: async () => {
            await AsyncStorage.removeItem('suach_onboarding_config_v1');
            onboardingState$.isCommitted.set(false);
            onboardingState$.currentStep.set(1);
            onboardingState$.secondsRemaining.set(10);
            onboardingState$.isHolding.set(false);
            router.replace('/');
          },
        },
      ]
    );
  };

  return (
    <ScrollView className="flex-1 bg-[#0A0A0A]" contentContainerStyle={{ paddingHorizontal: 24, paddingTop: 60, paddingBottom: 40, gap: 20 }}>
      <View className="mb-2">
        <Text className="text-3xl font-black text-white">Settings & Calibration</Text>
        <Text className="text-sm text-[#7CA08C] mt-1">Your personal Suach parameters</Text>
      </View>

      {/* Tribal Identity Card */}
      <View className="bg-[#121212] rounded-2xl p-5 border border-[#222222]">
        <Text className="text-[10px] font-extrabold text-[#3EF090] tracking-widest mb-3.5 uppercase">
          ACTIVE TRIBAL IDENTITY
        </Text>
        <View className="flex-row items-center justify-between mb-3">
          <View className="bg-[#1C3226] px-3 py-1.5 rounded-lg">
            <Text className="text-xs font-black text-[#3EF090]">
              {currentTribalInfo.classGroup}
            </Text>
          </View>
          <Text className="text-xs font-extrabold text-[#E0E0E0]">
            {consistencyTarget}% Target
          </Text>
        </View>

        <View className="bg-[#0E1A14] p-3 rounded-xl border-l-4 border-l-[#3EF090] my-2.5">
          <Text className="text-xs italic text-[#C0E0D0] leading-5">
            {currentTribalInfo.scripture}
          </Text>
        </View>

        <Text className="text-xs leading-5 text-[#999999]">
          {currentTribalInfo.vision}
        </Text>
      </View>

      {/* Daily Touchpoints Card */}
      <View className="bg-[#121212] rounded-2xl p-5 border border-[#222222]">
        <Text className="text-[10px] font-extrabold text-[#3EF090] tracking-widest mb-3.5 uppercase">
          DAILY FOCUS TOUCHPOINTS
        </Text>

        <View className="flex-row justify-between items-center py-1">
          <Text className="text-xs text-[#888888] font-semibold">Morning Wisdom:</Text>
          <Text className="text-sm font-extrabold text-[#3EF090]">{morningTime}</Text>
        </View>

        <View className="h-px bg-[#1F1F1F] my-2.5" />

        <View className="flex-row justify-between items-center py-1">
          <Text className="text-xs text-[#888888] font-semibold">Evening Rest:</Text>
          <Text className="text-sm font-extrabold text-[#3EF090]">{eveningTime}</Text>
        </View>

        <View className="h-px bg-[#1F1F1F] my-2.5" />

        <View className="flex-row justify-between items-center py-1">
          <Text className="text-xs text-[#888888] font-semibold">Initial Diagnosis:</Text>
          <Text className="text-xs font-semibold text-[#CCCCCC] max-w-45 text-right">
            {userDiagnosis ?? 'N/A'}
          </Text>
        </View>
      </View>

      {/* Re-calibration Action */}
      <Pressable
        className="bg-[#1A1A1A] rounded-2xl py-4 items-center border border-[#333333] mt-2.5 active:opacity-80"
        onPress={handleResetOnboarding}
      >
        <Text className="text-sm font-extrabold text-[#FF5555]">
          Re-Calibrate Onboarding
        </Text>
      </Pressable>
    </ScrollView>
  );
});
