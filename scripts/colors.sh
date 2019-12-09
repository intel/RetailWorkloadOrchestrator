#!/bin/bash
textutils="textutils.sh"
source "${textutils}"

# Test
echo -e "${T_BOLD}${C_RED}RED${C_GREEN}GREEN${C_YELLOW}YELLOW${C_BLUE}BLUE"
echo -e "${C_MAGENTA}PURPLE${C_CYAN}CYAN${C_WHITE}WHITE"
echo -e "${C_GRAY}LIGHTGRAY${C_L_RED}LightRed${C_L_GREEN}LGREEN${C_L_YELLOW}LYELLOW"
echo -e "${C_L_BLUE}LBLUE${C_L_MAGENTA}LMAGENTA${C_L_CYAN}LCYAN${C_L_WHITE}LWHITE ${T_RESET}"
echo -e "${C_GRAY}this is not too important${T_RESET}"
echo -e "${T_ERR}The file was not found${T_RESET}"
printMsg "${T_INFO_ICON} No results found for ${C_BOLD}Query${T_RESET}"