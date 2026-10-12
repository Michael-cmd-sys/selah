import React, { useState } from 'react';
import { View, ActivityIndicator } from 'react-native';
import { Redirect } from 'expo-router';
import { observer } from '@legendapp/state/react';
import { OnboardingContainer } from '@/features/onboarding/OnboardingContainer';
import {
  onboardingState$,
  loadSavedOnboardingState,
} from '@/features/onboarding/onboardingState';

export default observer(function OnboardingScreen() {
  const [isLoading, setIsLoading] = useState(true);
  const isCommitted = onboardingState$.isCommitted.get();
  const initStarted = React.useRef(false);

  if (!initStarted.current) {
    initStarted.current = true;
    loadSavedOnboardingState().then(() => {
      setIsLoading(false);
    });
  }

  if (isLoading) {
    return (
      <View className="flex-1 bg-[#0A0A0A] items-center justify-center">
        <ActivityIndicator size="large" color="#3EF090" />
      </View>
    );
  }

  if (isCommitted) {
    return <Redirect href="/(tabs)" />;
  }

  return <OnboardingContainer />;
});
