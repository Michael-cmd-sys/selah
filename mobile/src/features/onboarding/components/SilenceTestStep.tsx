import React, { useRef, useState } from 'react';
import { View, Text, Pressable, Animated } from 'react-native';
import { observer } from '@legendapp/state/react';
import * as Haptics from 'expo-haptics';
import Svg, { Circle } from 'react-native-svg';
import { onboardingState$ } from '../onboardingState';

// Legend-State isolated reactive countdown component
const IsolatedTimerText = observer(() => {
  const seconds = onboardingState$.secondsRemaining.get();
  const holding = onboardingState$.isHolding.get();

  return (
    <View className="items-center justify-center">
      <Text className="text-5xl font-black text-[#3EF090] font-mono">
        {seconds}
      </Text>
      <Text className="text-[9px] font-extrabold text-[#888888] tracking-[1.5px] mt-1 uppercase">
        {holding ? 'SEC OF SILENCE' : 'SECONDS HOLD'}
      </Text>
    </View>
  );
});

// Legend-State isolated warning label component
const IsolatedWarningText = observer(({ showWarning }: { showWarning: boolean }) => {
  if (!showWarning) return <View className="h-6 mt-5" />;
  return (
    <Text className="mt-5 text-xs font-semibold text-[#FF5555] tracking-wide h-6 text-center">
      Stay still... try again.
    </Text>
  );
});

export const SilenceTestStep: React.FC = () => {
  const timerRef = useRef<| null>(null);
  const [showWarning, setShowWarning] = useState(false);

  const pulseAnim = useRef(new Animated.Value(1)).current;
  const scaleAnim = useRef(new Animated.Value(1)).current;

  const startTimer = () => {
    setShowWarning(false);
    onboardingState$.isHolding.set(true);
    onboardingState$.secondsRemaining.set(10);

    Haptics.impactAsync(Haptics.ImpactFeedbackStyle.Medium);

    Animated.loop(
      Animated.sequence([
        Animated.timing(pulseAnim, {
          toValue: 1.12,
          duration: 600,
          useNativeDriver: true,
        }),
        Animated.timing(pulseAnim, {
          toValue: 1,
          duration: 600,
          useNativeDriver: true,
        }),
      ])
    ).start();

    Animated.spring(scaleAnim, {
      toValue: 0.94,
      useNativeDriver: true,
    }).start();

    if (timerRef.current) clearInterval(timerRef.current);

    timerRef.current = setInterval(() => {
      const current = onboardingState$.secondsRemaining.get();
      if (current > 1) {
        onboardingState$.secondsRemaining.set(current - 1);
        Haptics.impactAsync(Haptics.ImpactFeedbackStyle.Light);
      } else if (current === 1) {
        onboardingState$.secondsRemaining.set(0);
        if (timerRef.current) {
          clearInterval(timerRef.current);
          timerRef.current = null;
        }

        Haptics.notificationAsync(Haptics.NotificationFeedbackType.Success);
        onboardingState$.isHolding.set(false);

        // Transition to Step 2
        setTimeout(() => {
          onboardingState$.currentStep.set(2);
        }, 300);
      }
    }, 1000);
  };

  const cancelTimer = () => {
    const remaining = onboardingState$.secondsRemaining.get();

    pulseAnim.stopAnimation();
    pulseAnim.setValue(1);
    Animated.spring(scaleAnim, {
      toValue: 1,
      useNativeDriver: true,
    }).start();

    if (timerRef.current) {
      clearInterval(timerRef.current);
      timerRef.current = null;
    }

    if (onboardingState$.isHolding.get() && remaining > 0) {
      onboardingState$.isHolding.set(false);
      onboardingState$.secondsRemaining.set(10);
      setShowWarning(true);
      Haptics.notificationAsync(Haptics.NotificationFeedbackType.Warning);
    }
  };

  return (
    <View className="flex-1 bg-[#0A0A0A] justify-between px-7 pt-12 pb-10">
      <View className="items-center mt-5">
        <Text className="text-[10px] font-extrabold text-[#3EF090] tracking-[2px] mb-3 uppercase">
          STEP 01 // DIAGNOSTIC SILENCE
        </Text>
        <Text className="text-3xl font-black text-white text-center tracking-tight mb-3">
          Welcome to Suach.
        </Text>
        <Text className="text-sm leading-5 text-[#A0A0A0] text-center px-3">
          Press and hold the circle below for 10 seconds of absolute silence. Do not lift your finger.
        </Text>
      </View>

      <View className="items-center justify-center h-[300px]">
        <Animated.View
          className="absolute w-[260px] h-[260px] rounded-full bg-[#3EF090]"
          style={{
            transform: [{ scale: pulseAnim }],
            opacity: onboardingState$.isHolding.get() ? 0.35 : 0.08,
          }}
        />

        <Pressable
          onPressIn={startTimer}
          onPressOut={cancelTimer}
          className="w-[220px] h-[220px] items-center justify-center"
        >
          <Animated.View
            className="w-[220px] h-[220px] rounded-full bg-[#111111] border border-[#2A2A2A] items-center justify-center shadow-lg shadow-[#3EF090]/20"
            style={{ transform: [{ scale: scaleAnim }] }}
          >
            <Svg width={220} height={220} className="absolute">
              <Circle
                cx="110"
                cy="110"
                r="100"
                stroke="#222222"
                strokeWidth="6"
                fill="none"
              />
              <Circle
                cx="110"
                cy="110"
                r="100"
                stroke="#3EF090"
                strokeWidth="6"
                fill="none"
                strokeDasharray="628"
                strokeDashoffset={
                  628 - (628 * (10 - onboardingState$.secondsRemaining.get())) / 10
                }
                strokeLinecap="round"
                transform="rotate(-90 110 110)"
              />
            </Svg>

            <IsolatedTimerText />
          </Animated.View>
        </Pressable>

        <IsolatedWarningText showWarning={showWarning} />
      </View>

      <Text className="text-center text-[11px] color-[#444444] tracking-widest uppercase font-medium">
        Somatic Focus Baseline • Zero Dopamine Loop
      </Text>
    </View>
  );
};
