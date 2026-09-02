import { Box, BoxProps } from "@mui/material";
import React, { PropsWithChildren } from "react";

export const Container: React.FC<
  PropsWithChildren<{ boxProps?: BoxProps }>
> = ({ children, boxProps = {} }) => (
  <Box
    {...boxProps}
    sx={{
      width: "100%",
      maxWidth: "752px",
      marginX: "auto",
      backgroundColor: "background.paper",
      padding: 3,
      borderRadius: 2,
      ...boxProps.sx,
    }}
  >
    {children}
  </Box>
);
