import { View, Text, StyleSheet } from "react-native";

export function BibleScreen() {
  return (
    <View style={styles.container}>
      <Text style={styles.title}>Bible</Text>
      <Text style={styles.subtitle}>Coming soon</Text>
    </View>
  );
}

const styles = StyleSheet.create({
  container: {
    flex: 1,
    backgroundColor: "#0D1C17",
    alignItems: "center",
    justifyContent: "center",
  },
  title: {
    fontSize: 24,
    fontWeight: "bold",
    color: "#ffffff",
  },
  subtitle: {
    fontSize: 16,
    color: "#7CA08C",
    marginTop: 8,
  },
});
