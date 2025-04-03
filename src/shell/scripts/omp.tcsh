setenv POWERLINE_COMMAND "oh-my-posh";
setenv POSH_THEME ::CONFIG::;
setenv POSH_SHELL "tcsh";
setenv POSH_SHELL_VERSION "$tcsh";
setenv POSH_SESSION_ID ::SESSION_ID::;
setenv OSTYPE "$OSTYPE";

setenv VIRTUAL_ENV_DISABLE_PROMPT 1;
setenv PYENV_VIRTUALENV_DISABLE_PROMPT 1;

if ( ! $?_omp_enabled ) alias precmd '
  set _omp_status = $status;
  set _omp_execution_time = -1;
  set _omp_last_cmd = `echo $_:q`;
  if ( $#_omp_last_cmd && $?_omp_cmd_executed ) @ _omp_execution_time = `"$_omp_executable" get millis` - $_omp_start_time;
  unset _omp_last_cmd;
  unset _omp_cmd_executed;
  @ _omp_stack_count = $#dirstack - 1;
  set prompt = "`$_omp_executable:q print primary
    --save-cache
    --shell=tcsh
    --shell-version=$tcsh
    --status=$_omp_status
    --execution-time=$_omp_execution_time
    --stack-count=$_omp_stack_count`";
'"`alias precmd`";

if ( ! $?_omp_enabled ) alias postcmd '
  set _omp_start_time = `"$_omp_executable" get millis`;
  set _omp_cmd_executed;
'"`alias postcmd`";

set _omp_enabled;
set _omp_executable = ::OMP::;
set _omp_execution_time = -1;
set _omp_start_time = -1;
