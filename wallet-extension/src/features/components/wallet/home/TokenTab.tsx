import type { Hex } from "@nilfoundation/niljs";
import { Button, COLORS, CopyButton, HeadingMedium, ParagraphSmall } from "@nilfoundation/ui-kit";
import { useStore } from "effector-react";
import { useState } from "react";
import { useTranslation } from "react-i18next";
import { useNavigate } from "react-router-dom";
import nilIcon from "../../../../../public/icons/token/nil.svg";
import { WalletRoutes } from "../../../../router";
import {
  $balance,
  $balanceToken,
  $tokens,
  getBalanceForToken,
} from "../../../store/model/token.ts";
import {
  convertWeiToEth,
  formatAddress,
  getTokenIcon,
  getTokens,
  isMockToken,
} from "../../../utils";
import { Box, Icon } from "../../shared";

export const TokenTab = () => {
  const balance = useStore($balance);
  const balanceToken = useStore($balanceToken);
  const allTokens = useStore($tokens);
  const [clicked, setClicked] = useState(false);
  const [hoveredIndex, setHoveredIndex] = useState<number | null>(null);
  const { t } = useTranslation("translation");
  const navigate = useNavigate();
  const availableTokens = getTokens(allTokens, true);

  const handleCopy = (address: string) => {
    navigator.clipboard.writeText(address).then(() => {
      console.log(`Address ${address} copied to clipboard!`);
    });
  };

  const tokens = availableTokens.map((token) => {
    const amount = getBalanceForToken(token.address, balance ?? 0n, balanceToken ?? {});
    if (token.address === "") {
      return {
        icon: nilIcon,
        title: "Nil",
        subtitle: "Native token",
        subtitleColor: COLORS.green200,
        rightText: balance !== null ? `${convertWeiToEth(balance)} NIL` : "0 NIL",
        rightTextColor: COLORS.gray50,
        address: "",
      };
    }
    return {
      icon: getTokenIcon(token.label),
      title: token.label,
      subtitle: isMockToken(token.label) ? "Mock token" : formatAddress(token.address as Hex),
      subtitleColor: "gray",
      rightText: `${amount.toString()} ${token.label}`,
      rightTextColor: COLORS.gray50,
      address: token.address,
    };
  });

  if (tokens.length === 0) {
    return (
      <Box $style={{ textAlign: "center", paddingTop: "40px" }}>
        <ParagraphSmall $style={{ color: COLORS.gray200 }}>No tokens found</ParagraphSmall>
      </Box>
    );
  }

  return (
    <Box
      $style={{
        paddingTop: "3px",
        display: "flex",
        flexDirection: "column",
        gap: "6px",
        maxHeight: "calc(100vh - 120px)",
        overflowY: "auto",
        "-ms-overflow-style": "none",
        "scrollbar-width": "none",
        height: "285px",
        margin: "0 8px",
      }}
    >
      {tokens.map((token, index) => (
        <Box
          key={`${token.title}-${index}`}
          $align="center"
          $justify="space-between"
          $style={{
            flexDirection: "row",
            width: "100%",
            padding: "5px",
            cursor: "pointer",
            backgroundColor: hoveredIndex === index ? COLORS.gray800 : "transparent",
            transition: "background-color 0.3s",
          }}
          onMouseEnter={() => setHoveredIndex(index)}
          onMouseLeave={() => setHoveredIndex(null)}
        >
          <Box
            $align="center"
            $gap="8px"
            $style={{
              flexDirection: "row",
            }}
          >
            <Icon
              src={token.icon}
              alt={`${token.title} Icon`}
              size={64}
              iconSize="100%"
              background="transparent"
            />
            <Box $align="flex-start" $style={{ flexDirection: "column", maxWidth: "200px" }}>
              <HeadingMedium
                $style={{
                  color: COLORS.gray50,
                  whiteSpace: "nowrap",
                  overflow: "hidden",
                  textOverflow: "ellipsis",
                }}
              >
                {token.title || "Custom Token"}
              </HeadingMedium>
              <ParagraphSmall
                onClick={() => {
                  if (token.title === "") {
                    handleCopy(token.address);
                    setClicked(true);
                    setTimeout(() => setClicked(false), 2000);
                  }
                }}
                $style={{
                  color: token.subtitleColor || COLORS.gray200,
                  whiteSpace: "nowrap",
                  cursor: token.title === "" ? "pointer" : "default",
                  transition: "color 0.3s",
                  ":hover": {
                    color: token.title === "" ? COLORS.gray300 : token.subtitleColor,
                  },
                }}
              >
                {hoveredIndex === index && token.address ? (
                  <>
                    {formatAddress(token.address as Hex)}
                    <CopyButton textToCopy={token.address} />
                  </>
                ) : clicked && token.title === "" ? (
                  "Copied"
                ) : (
                  token.subtitle
                )}
              </ParagraphSmall>
            </Box>
          </Box>
          <Box
            $align="center"
            $justify="flex-end"
            $gap="8px"
            $style={{
              flexDirection: "row",
            }}
          >
            <ParagraphSmall
              $style={{
                color: token.rightTextColor || COLORS.gray50,
                textAlign: "right",
                whiteSpace: "nowrap",
              }}
            >
              {token.rightText}
            </ParagraphSmall>
          </Box>
        </Box>
      ))}
      <Box>
        <Button
          onClick={() => {
            navigate(WalletRoutes.WALLET.MANAGE_TOKENS);
          }}
          overrides={{
            Root: {
              style: {
                width: "100%",
                height: "48px",
                backgroundColor: COLORS.gray800,
                color: COLORS.gray200,
                ":hover": {
                  backgroundColor: COLORS.gray700,
                },
              },
            },
          }}
        >
          {t("wallet.manageTokens.tokenTab.button")}
        </Button>
      </Box>
    </Box>
  );
};
