import React, { useRef } from 'react';
import { View, Text, Pressable, Animated, ScrollView } from 'react-native';
import { observer } from '@legendapp/state/react';
import * as Haptics from 'expo-haptics';
import { onboardingState$ } from '../onboardingState';

const POEM_BLOCKS = [
  {
    delay: 0,
    lines: [
      "Whatever the cause may be",
      "I'll fight through it all",
      "Just to keep on moving forward",
      "I'll keep trying to keep my stand.",
    ],
  },
  {
    delay: 1500,
    lines: [
      "For my God is my vine",
      "And I am the branches",
      "I need to abide in Him to keep on flourishing",
      "So I abide in Him to get my nutrients to grow.",
    ],
  },
  {
    delay: 3000,
    lines: [
      "He's my Shepherd. He's my guide. He's my friend.",
      "So I won't fear what the enemy brings my way",
      "All I say is bring it on.",
    ],
  },
];

export const NarrativeStep: React.FC = observer(() => {
  const anims = useRef(POEM_BLOCKS.map(() => new Animated.Value(0))).current;
  const translates = useRef(POEM_BLOCKS.map(() => new Animated.Value(20))).current;
  const animationStarted = useRef(false);

  // Lazy one-time animation launch without useEffect churn
  if (!animationStarted.current) {
    animationStarted.current = true;
    POEM_BLOCKS.forEach((block, index) => {
      setTimeout(() => {
        Haptics.impactAsync(Haptics.ImpactFeedbackStyle.Light);
        Animated.parallel([
          Animated.timing(anims[index], {
            toValue: 1,
            duration: 900,
            useNativeDriver: true,
          }),
          Animated.timing(translates[index], {
            toValue: 0,
            duration: 900,
            useNativeDriver: true,
          }),
        ]).start();
      }, block.delay);
    });
  }

  const handleProceed = () => {
    Haptics.impactAsync(Haptics.ImpactFeedbackStyle.Medium);
    onboardingState$.currentStep.set(5);
  };

  return (
    <View className="flex-1 bg-[#0A0A0A]">
      <ScrollView
        showsVerticalScrollIndicator={false}
        contentContainerStyle={{ paddingHorizontal: 24, paddingTop: 40, paddingBottom: 110 }}
      >
        <View className="mb-6">
          <Text className="text-[10px] font-extrabold text-[#3EF090] tracking-[2px] mb-2 uppercase">
            STEP 04 // FOUNDATIONAL PHILOSOPHY
          </Text>
          <Text className="text-3xl font-black text-white tracking-tight mb-2">
            The Creator's Poem
          </Text>
          <Text className="text-sm leading-5 text-[#999999]">
            The spiritual anchor and operating ethos behind Suach.
          </Text>
        </View>

        <View className="gap-5">
          {POEM_BLOCKS.map((block, i) => (
            <Animated.View
              key={i}
              className="bg-[#121212] rounded-2xl p-5 border border-[#222222] flex-row items-start gap-4"
              style={{
                opacity: anims[i],
                transform: [{ translateY: translates[i] }],
              }}
            >
              <View className="w-7 h-7 rounded-lg bg-[#1C2E24] items-center justify-center mt-0.5">
                <Text className="text-xs font-black text-[#3EF090]">0{i + 1}</Text>
              </View>
              <View className="flex-1 gap-1.5">
                {block.lines.map((line, lIndex) => (
                  <Text key={lIndex} className="text-base leading-6 text-[#E2E2E2] font-medium italic">
                    {line}
                  </Text>
                ))}
              </View>
            </Animated.View>
          ))}
        </View>
      </ScrollView>

      {/* Bottom CTA Button */}
      <View className="absolute bottom-0 left-0 right-0 bg-[#0A0A0A] px-6 py-4 border-t border-t-[#1A1A1A]">
        <Pressable
          className="flex-row bg-[#3EF090] rounded-2xl py-4 items-center justify-center gap-2 active:opacity-85"
          onPress={handleProceed}
        >
          <Text className="text-base font-extrabold text-[#0A0A0A]">
            Proceed to Covenant
          </Text>
          <Text className="text-lg font-extrabold text-[#0A0A0A]">→</Text>
        </Pressable>
      </View>
    </View>
  );
});
