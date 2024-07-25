setenv POWERLINE_COMMAND "oh-my-posh";
setenv POSH_THEME ::CONFIG::;
setenv POSH_SHELL_VERSION "";
setenv POSH_PID $$;

set POSH_COMMAND = ::OMP::;
set USER_PRECMD = "`alias precmd`";
set USER_POSTCMD = "`alias postcmd`";
set POSH_PRECMD = 'set POSH_CMD_STATUS = $status;set POSH_END_TIME = `$POSH_COMMAND get millis`;set POSH_DURATION = 0;if ( $POSH_START_TIME != -1 ) @ POSH_DURATION = $POSH_END_TIME - $POSH_START_TIME;set prompt = "`$POSH_COMMAND print primary --shell=tcsh --status=$POSH_CMD_STATUS --execution-time=$POSH_DURATION`";set POSH_START_TIME = -1';
set POSH_POSTCMD = 'set POSH_START_TIME = `$POSH_COMMAND get millis`';

alias precmd "$POSH_PRECMD;$USER_PRECMD";
alias postcmd "$POSH_POSTCMD;$USER_POSTCMD";

set POSH_START_TIME = `$POSH_COMMAND get millis`;
