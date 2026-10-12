import React, { useState } from 'react';
import { View, Text, Pressable, ScrollView, Modal } from 'react-native';
import { observer } from '@legendapp/state/react';
import * as Haptics from 'expo-haptics';
import { onboardingState$, TRIBAL_CLASSES } from '../onboardingState';

const MORNING_OPTIONS = ['05:00 AM', '05:30 AM', '06:00 AM', '06:30 AM', '07:00 AM', '07:30 AM', '08:00 AM', '08:30 AM'];
const EVENING_OPTIONS = ['08:00 PM', '08:30 PM', '09:00 PM', '09:30 PM', '10:00 PM', '10:30 PM', '11:00 PM', '11:30 PM'];
const TARGET_PERCENTAGES = [50, 70, 85, 95, 100];

export const TimeAndTribeStep: React.FC = observer(() => {
  const morningTime = onboardingState$.morningTime.get();
  const eveningTime = onboardingState$.eveningTime.get();
  const consistencyTarget = onboardingState$.consistencyTarget.get();

  const currentTribalInfo = TRIBAL_CLASSES[consistencyTarget] ?? TRIBAL_CLASSES[70];

  const [activeModal, setActiveModal] = useState<'morning' | 'evening' | null>(null);

  const handleSelectTarget = (pct: number) => {
    Haptics.selectionAsync();
    onboardingState$.consistencyTarget.set(pct);
    const info = TRIBAL_CLASSES[pct];
    if (info) {
      const className = info.classGroup.replace('CLASS: ', '');
      onboardingState$.tribalClass.set(className);
    }
  };

  const handleSelectMorning = (time: string) => {
    Haptics.selectionAsync();
    onboardingState$.morningTime.set(time);
    setActiveModal(null);
  };

  const handleSelectEvening = (time: string) => {
    Haptics.selectionAsync();
    onboardingState$.eveningTime.set(time);
    setActiveModal(null);
  };

  const handleProceed = () => {
    Haptics.impactAsync(Haptics.ImpactFeedbackStyle.Medium);
    onboardingState$.currentStep.set(4);
  };

  return (
    <View className="flex-1 bg-[#0A0A0A]">
      <ScrollView
        showsVerticalScrollIndicator={false}
        contentContainerStyle={{ paddingHorizontal: 24, paddingTop: 40, paddingBottom: 110 }}
      >
        <View className="mb-6">
          <Text className="text-[10px] font-extrabold text-[#3EF090] tracking-[2px] mb-2 uppercase">
            STEP 03 // PERSONAL RE-CALIBRATION
          </Text>
          <Text className="text-3xl font-black text-white tracking-tight mb-2">
            Time & Tribe
          </Text>
          <Text className="text-sm leading-5 text-[#999999]">
            Set your daily focus windows and choose a grace-inclusive consistency target.
          </Text>
        </View>

        {/* Section 1: Dual Time Pickers */}
        <View className="bg-[#121212] rounded-2xl p-4.5 border border-[#222222] mb-6">
          <Text className="text-[10px] font-extrabold text-[#777777] tracking-widest mb-3.5 uppercase">
            DAILY FOCUS WINDOWS
          </Text>

          {/* Morning Picker */}
          <View className="flex-row items-center justify-between py-1.5">
            <View className="flex-1">
              <Text className="text-base font-bold text-[#EEEEEE]">Morning Wisdom</Text>
              <Text className="text-xs text-[#666666] mt-0.5">Morning Proverbs touchpoint</Text>
            </View>
            <Pressable
              className="flex-row items-center bg-[#1C1C1C] px-3.5 py-2.5 rounded-xl border border-[#333333] gap-2"
              onPress={() => setActiveModal('morning')}
            >
              <Text className="text-sm font-bold text-[#3EF090]">{morningTime}</Text>
              <Text className="text-xs text-[#888888]">▾</Text>
            </Pressable>
          </View>

          <View className="h-px bg-[#1F1F1F] my-3" />

          {/* Evening Picker */}
          <View className="flex-row items-center justify-between py-1.5">
            <View className="flex-1">
              <Text className="text-base font-bold text-[#EEEEEE]">Evening Soul Rest</Text>
              <Text className="text-xs text-[#666666] mt-0.5">Evening Psalms decompress</Text>
            </View>
            <Pressable
              className="flex-row items-center bg-[#1C1C1C] px-3.5 py-2.5 rounded-xl border border-[#333333] gap-2"
              onPress={() => setActiveModal('evening')}
            >
              <Text className="text-sm font-bold text-[#3EF090]">{eveningTime}</Text>
              <Text className="text-xs text-[#888888]">▾</Text>
            </Pressable>
          </View>
        </View>

        {/* Section 2: Segmented Target Selector */}
        <View className="mb-6">
          <Text className="text-[10px] font-extrabold text-[#777777] tracking-widest mb-3.5 uppercase">
            ROLLING MONTHLY CONSISTENCY TARGET
          </Text>

          <View className="flex-row bg-[#121212] rounded-xl p-1 border border-[#222222] gap-1">
            {TARGET_PERCENTAGES.map((pct) => {
              const isSelected = consistencyTarget === pct;
              return (
                <Pressable
                  key={pct}
                  className={`flex-1 py-3 items-center rounded-lg ${isSelected ? 'bg-[#3EF090]' : 'bg-transparent'
                    }`}
                  onPress={() => handleSelectTarget(pct)}
                >
                  <Text
                    className={`text-xs font-bold ${isSelected ? 'text-[#0A0A0A] font-extrabold' : 'text-[#777777]'
                      }`}
                  >
                    {pct}%
                  </Text>
                </Pressable>
              );
            })}
          </View>
        </View>

        {/* Section 3: Dynamic Tribal Anchor Card Engine */}
        <View className="bg-[#131A16] rounded-2xl p-5 border border-[#274535]">
          <View className="flex-row items-center justify-between mb-3.5">
            <View className="bg-[#1E3B2C] px-3 py-1.5 rounded-lg">
              <Text className="text-[11px] font-black text-[#3EF090] tracking-wider uppercase">
                {currentTribalInfo.classGroup}
              </Text>
            </View>
            <Text className="text-xs font-bold text-[#88C2A2]">
              {currentTribalInfo.targetDays}
            </Text>
          </View>

          <View className="bg-[#0B130E] rounded-xl p-3.5 border-l-4 border-l-[#3EF090] mb-3.5">
            <Text className="text-xs italic text-[#D0E8DB] leading-5">
              {currentTribalInfo.scripture}
            </Text>
          </View>

          <Text className="text-xs leading-5 text-[#A0B8AA]">
            {currentTribalInfo.vision}
          </Text>
        </View>
      </ScrollView>

      {/* Bottom CTA Button */}
      <View className="absolute bottom-0 left-0 right-0 bg-[#0A0A0A] px-6 py-4 border-t border-t-[#1A1A1A]">
        <Pressable
          className="flex-row bg-[#3EF090] rounded-2xl py-4 items-center justify-center gap-2 active:opacity-85"
          onPress={handleProceed}
        >
          <Text className="text-base font-extrabold text-[#0A0A0A]">
            Proceed to Foundation
          </Text>
          <Text className="text-lg font-extrabold text-[#0A0A0A]">→</Text>
        </Pressable>
      </View>

      {/* Time Selector Modal */}
      <Modal
        visible={activeModal !== null}
        transparent
        animationType="fade"
        onRequestClose={() => setActiveModal(null)}
      >
        <Pressable
          className="flex-1 bg-black/85 justify-center items-center p-6"
          onPress={() => setActiveModal(null)}
        >
          <View className="w-full bg-[#141414] rounded-2xl p-6 border border-[#2A2A2A]">
            <Text className="text-lg font-extrabold text-white mb-4.5 text-center">
              {activeModal === 'morning' ? 'Select Morning Time' : 'Select Evening Time'}
            </Text>

            <View className="flex-row flex-wrap justify-between gap-2.5 mb-5">
              {(activeModal === 'morning' ? MORNING_OPTIONS : EVENING_OPTIONS).map((t) => {
                const isSelected = (activeModal === 'morning' ? morningTime : eveningTime) === t;
                return (
                  <Pressable
                    key={t}
                    className={`w-[48%] py-3 rounded-xl items-center border ${isSelected
                        ? 'bg-[#1B3526] border-[#3EF090]'
                        : 'bg-[#1E1E1E] border-[#2D2D2D]'
                      }`}
                    onPress={() =>
                      activeModal === 'morning' ? handleSelectMorning(t) : handleSelectEvening(t)
                    }
                  >
                    <Text
                      className={`text-sm font-bold ${isSelected ? 'text-[#3EF090]' : 'text-[#CCCCCC]'
                        }`}
                    >
                      {t}
                    </Text>
                  </Pressable>
                );
              })}
            </View>

            <Pressable
              className="bg-[#2A2A2A] py-3.5 rounded-xl items-center"
              onPress={() => setActiveModal(null)}
            >
              <Text className="text-sm font-bold text-white">Done</Text>
            </Pressable>
          </View>
        </Pressable>
      </Modal>
    </View>
  );
});
