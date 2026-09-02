import React from "react";

import { Account } from "../types";
import { Box, Stack, Typography } from "@mui/material";

interface AccountDetailsProps {
  account: Account | undefined;
}

const AccountDetails: React.FC<AccountDetailsProps> = ({ account }) => {
  return (
    <Box sx={{ marginY: 4 }}>
      <Stack spacing={1}>
        <Stack flexDirection="row">
          <Typography color="secondary" sx={{ mr: 1 }}>
            Address:
          </Typography>
          <Typography color="text.primary">{account?.address}</Typography>
        </Stack>
      </Stack>
      <Stack flexDirection="row">
        <Typography color="secondary" sx={{ mr: 1 }}>
          Public Key:
        </Typography>
        <Typography color="text.primary">{account?.pubKey}</Typography>
      </Stack>
      <Stack flexDirection="row">
        <Typography color="secondary" sx={{ mr: 1 }}>
          HD Path:
        </Typography>
        <Typography color="text.primary">{account?.hdPath}</Typography>
      </Stack>
    </Box>
  );
};

export default AccountDetails;
