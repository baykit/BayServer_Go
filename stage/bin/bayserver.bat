@ECHO OFF

REM 
REM  BayServer boot script
REM 

set daemon=0
for %%f in (%*) do (
  if "%%f"=="-daemon" (
     set daemon=1
  )
)

if "%daemon%" == "1" (
  start %~p0\bayserver %*
) else (
  %~p0\bayserver %*

)
