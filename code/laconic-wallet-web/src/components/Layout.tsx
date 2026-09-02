import { Button, Typography } from "@mui/material";
import React, { PropsWithChildren } from "react";
import { Container } from "./Container";
import { ArrowBack } from "@mui/icons-material";
import { NativeStackNavigationProp } from "@react-navigation/native-stack";
import { useNavigation } from "@react-navigation/native";
import { StackParamsList } from "../types";

export const Layout: React.FC<PropsWithChildren<{ title: string }>> = ({
  children,
  title,
}) => {
  const navigation =
    useNavigation<NativeStackNavigationProp<StackParamsList>>();

  return (
    <Container
      boxProps={{ sx: { backgroundColor: "inherit", padding: 0, mt: 3 } }}
    >
      <Button
        startIcon={<ArrowBack />}
        color="info"
        sx={{ mb: 4 }}
        onClick={() => navigation.navigate("Home")}
      >
        Home
      </Button>
      <Typography variant="h4" sx={{ mb: 4 }}>
        {title}
      </Typography>
      <Container>{children}</Container>
    </Container>
  );
};
