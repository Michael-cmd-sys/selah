import React, { useRef, useState } from 'react';
import {
  View,
  Text,
  PanResponder,
  Animated,
  Dimensions,
} from 'react-native';
import { observer } from '@legendapp/state/react';
import * as Haptics from 'expo-haptics';
import { useRouter } from 'expo-router';
import { onboardingState$, saveOnboardingStateToStorage } from '../onboardingState';
import { scheduleSuachNotifications } from '../utils/notifications';

const { width: SCREEN_WIDTH } = Dimensions.get('window');
const TRACK_PADDING = 6;
const TRACK_WIDTH = SCREEN_WIDTH - 48;
const THUMB_SIZE = 56;
const MAX_SLIDE = TRACK_WIDTH - THUMB_SIZE - TRACK_PADDING * 2;

export const CovenantLockStep: React.FC = observer(() => {
  const router = useRouter();
  const consistencyTarget = onboardingState$.consistencyTarget.get();
  const tribalClass = onboardingState$.tribalClass.get();
  const morningTime = onboardingState$.morningTime.get();
  const eveningTime = onboardingState$.eveningTime.get();

  const panX = useRef(new Animated.Value(0)).current;
  const [isLockedIn, setIsLockedIn] = useState(false);

  const handleCommitmentSuccess = async () => {
    if (isLockedIn) return;
    setIsLockedIn(true);

    Haptics.notificationAsync(Haptics.NotificationFeedbackType.Success);

    await saveOnboardingStateToStorage();
    await scheduleSuachNotifications(morningTime, eveningTime);

    setTimeout(() => {
      router.replace('/(tabs)');
    }, 500);
  };

  const panResponder = useRef(
    PanResponder.create({
      onStartShouldSetPanResponder: () => true,
      onMoveShouldSetPanResponder: () => true,
      onPanResponderGrant: () => {
        Haptics.impactAsync(Haptics.ImpactFeedbackStyle.Light);
      },
      onPanResponderMove: (_, gestureState) => {
        if (isLockedIn) return;
        const newX = Math.max(0, Math.min(gestureState.dx, MAX_SLIDE));
        panX.setValue(newX);

        if (newX > MAX_SLIDE * 0.5 && newX < MAX_SLIDE * 0.55) {
          Haptics.impactAsync(Haptics.ImpactFeedbackStyle.Light);
        }
      },
      onPanResponderRelease: (_, gestureState) => {
        if (isLockedIn) return;

        if (gestureState.dx >= MAX_SLIDE * 0.75) {
          Animated.spring(panX, {
            toValue: MAX_SLIDE,
            useNativeDriver: false,
            bounciness: 6,
          }).start(() => {
            handleCommitmentSuccess();
          });
        } else {
          Animated.spring(panX, {
            toValue: 0,
            useNativeDriver: false,
            friction: 6,
          }).start();
        }
      },
    })
  ).current;

  const fillWidth = panX.interpolate({
    inputRange: [0, MAX_SLIDE],
    outputRange: [THUMB_SIZE, TRACK_WIDTH - TRACK_PADDING * 2],
    extrapolate: 'clamp',
  });

  const textOpacity = panX.interpolate({
    inputRange: [0, MAX_SLIDE * 0.6],
    outputRange: [1, 0],
    extrapolate: 'clamp',
  });

  return (
    <View className="flex-1 bg-[#0A0A0A] justify-between px-6 pt-12 pb-10">
      <View className="mt-2">
        <Text className="text-[10px] font-extrabold text-[#3EF090] tracking-[2px] mb-2 uppercase">
          STEP 05 // COVENANT LOCK
        </Text>
        <Text className="text-3xl font-black text-white tracking-tight">
          The Suach Commitment
        </Text>
      </View>

      <View className="bg-[#121212] rounded-2xl p-5 border border-[#222222]">
        <Text className="text-base leading-6 text-[#CCCCCC] mb-5">
          Suach is an ad-free, serverless sanctuary built for raw consistency. You have selected a target of{' '}
          <Text className="font-black text-[#3EF090]">{consistencyTarget}% consistency</Text>, entering the tribe of{' '}
          <Text className="font-black text-[#3EF090]">{tribalClass}</Text>. Are you ready to actively commit to drawing your daily nutrients?
        </Text>

        <View className="bg-[#1A1A1A] rounded-xl p-3.5 gap-2 border border-[#2A2A2A]">
          <View className="flex-row justify-between items-center">
            <Text className="text-xs font-semibold text-[#888888]">Morning Wisdom:</Text>
            <Text className="text-xs font-extrabold text-white">{morningTime}</Text>
          </View>
          <View className="flex-row justify-between items-center">
            <Text className="text-xs font-semibold text-[#888888]">Evening Rest:</Text>
            <Text className="text-xs font-extrabold text-white">{eveningTime}</Text>
          </View>
          <View className="flex-row justify-between items-center">
            <Text className="text-xs font-semibold text-[#888888]">Tribal Identity:</Text>
            <Text className="text-xs font-extrabold text-white">{tribalClass}</Text>
          </View>
        </View>
      </View>

      {/* Tactile Swipe Component / Covenant Lock Slider */}
      <View className="items-center">
        <View
          style={{ width: TRACK_WIDTH }}
          className="h-[68px] bg-[#141414] rounded-full border-1.5 border-[#2A2A2A] p-1.5 justify-center overflow-hidden"
        >
          {/* Active Fill Track */}
          <Animated.View
            className="absolute left-0 top-0 bottom-0 bg-[#1A3828] rounded-full"
            style={{ width: fillWidth }}
          />

          {/* Text Overlay */}
          <Animated.View
            className="absolute left-0 right-0 items-center justify-center"
            style={{ opacity: textOpacity }}
          >
            <Text className="text-sm font-extrabold text-[#3EF090] tracking-widest uppercase">
              {isLockedIn ? 'COMMITTED' : 'Swipe to Commit ››'}
            </Text>
          </Animated.View>

          {/* Draggable Thumb Button */}
          <Animated.View
            className={`w-[56px] h-[56px] rounded-full items-center justify-center shadow-lg shadow-[#3EF090]/40 ${
              isLockedIn ? 'bg-[#0A0A0A] border-2 border-[#3EF090]' : 'bg-[#3EF090]'
            }`}
            style={{ transform: [{ translateX: panX }] }}
            {...panResponder.panHandlers}
          >
            <Text className="text-xl font-black text-[#0A0A0A]">
              {isLockedIn ? '✓' : '➔'}
            </Text>
          </Animated.View>
        </View>

        <Text className="text-[11px] text-[#666666] mt-3 text-center">
          Swipe completely across to finalize your sacred agreement.
        </Text>
      </View>

      <View className="items-center">
        <Text className="text-[11px] text-[#444444] tracking-widest uppercase font-medium">
          Zero Ad Footprint • 100% Local Storage
        </Text>
      </View>
    </View>
  );
});
