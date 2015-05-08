##
## Auto Generated makefile by CodeLite IDE
## any manual changes will be erased      
##
## Release
ProjectName            :=usbint1
ConfigurationName      :=Release
WorkspacePath          := "D:\work\go\src\github.com\nodtem66\usbint1\c_code"
ProjectPath            := "D:\work\go\src\github.com\nodtem66\usbint1\c_code\usbint1"
IntermediateDirectory  :=./Release
OutDir                 := $(IntermediateDirectory)
CurrentFileName        :=
CurrentFilePath        :=
CurrentFileFullPath    :=
User                   :=jirawat
Date                   :=05/08/15
CodeLitePath           :="C:\Program Files (x86)\CodeLite"
LinkerName             :=D:/bin/TDM-GCC-64/bin/g++.exe 
SharedObjectLinkerName :=D:/bin/TDM-GCC-64/bin/g++.exe -shared -fPIC
ObjectSuffix           :=.o
DependSuffix           :=.o.d
PreprocessSuffix       :=.i
DebugSwitch            :=-g 
IncludeSwitch          :=-I
LibrarySwitch          :=-l
OutputSwitch           :=-o 
LibraryPathSwitch      :=-L
PreprocessorSwitch     :=-D
SourceSwitch           :=-c 
OutputFile             :=$(IntermediateDirectory)/$(ProjectName)
Preprocessors          :=$(PreprocessorSwitch)NDEBUG 
ObjectSwitch           :=-o 
ArchiveOutputSwitch    := 
PreprocessOnlySwitch   :=-E
ObjectsFileList        :="usbint1.txt"
PCHCompileFlags        :=
MakeDirCommand         :=makedir
RcCmpOptions           := 
RcCompilerName         :=D:/bin/TDM-GCC-64/bin/windres.exe 
LinkOptions            :=  -lusb-1.0 -L./lib64/
IncludePath            :=  $(IncludeSwitch). $(IncludeSwitch). 
IncludePCH             := 
RcIncludePath          := 
Libs                   := 
ArLibs                 :=  
LibPath                := $(LibraryPathSwitch). 

##
## Common variables
## AR, CXX, CC, AS, CXXFLAGS and CFLAGS can be overriden using an environment variables
##
AR       := D:/bin/TDM-GCC-64/bin/ar.exe rcu
CXX      := D:/bin/TDM-GCC-64/bin/g++.exe 
CC       := D:/bin/TDM-GCC-64/bin/gcc.exe 
CXXFLAGS :=  -O2 -Wall $(Preprocessors)
CFLAGS   :=  -O2 -Wall $(Preprocessors)
ASFLAGS  := 
AS       := D:/bin/TDM-GCC-64/bin/as.exe 


##
## User defined environment variables
##
CodeLiteDir:=C:\Program Files (x86)\CodeLite
Objects0=$(IntermediateDirectory)/main.c$(ObjectSuffix) $(IntermediateDirectory)/sqlite3.c$(ObjectSuffix) 



Objects=$(Objects0) 

##
## Main Build Targets 
##
.PHONY: all clean PreBuild PrePreBuild PostBuild
all: $(OutputFile)

$(OutputFile): $(IntermediateDirectory)/.d $(Objects) 
	@$(MakeDirCommand) $(@D)
	@echo "" > $(IntermediateDirectory)/.d
	@echo $(Objects0)  > $(ObjectsFileList)
	$(LinkerName) $(OutputSwitch)$(OutputFile) @$(ObjectsFileList) $(LibPath) $(Libs) $(LinkOptions)

$(IntermediateDirectory)/.d:
	@$(MakeDirCommand) "./Release"

PreBuild:


##
## Objects
##
$(IntermediateDirectory)/main.c$(ObjectSuffix): main.c $(IntermediateDirectory)/main.c$(DependSuffix)
	$(CC) $(SourceSwitch) "D:/work/go/src/github.com/nodtem66/usbint1/c_code/usbint1/main.c" $(CFLAGS) $(ObjectSwitch)$(IntermediateDirectory)/main.c$(ObjectSuffix) $(IncludePath)
$(IntermediateDirectory)/main.c$(DependSuffix): main.c
	@$(CC) $(CFLAGS) $(IncludePath) -MG -MP -MT$(IntermediateDirectory)/main.c$(ObjectSuffix) -MF$(IntermediateDirectory)/main.c$(DependSuffix) -MM "main.c"

$(IntermediateDirectory)/main.c$(PreprocessSuffix): main.c
	@$(CC) $(CFLAGS) $(IncludePath) $(PreprocessOnlySwitch) $(OutputSwitch) $(IntermediateDirectory)/main.c$(PreprocessSuffix) "main.c"

$(IntermediateDirectory)/sqlite3.c$(ObjectSuffix): sqlite3.c $(IntermediateDirectory)/sqlite3.c$(DependSuffix)
	$(CC) $(SourceSwitch) "D:/work/go/src/github.com/nodtem66/usbint1/c_code/usbint1/sqlite3.c" $(CFLAGS) $(ObjectSwitch)$(IntermediateDirectory)/sqlite3.c$(ObjectSuffix) $(IncludePath)
$(IntermediateDirectory)/sqlite3.c$(DependSuffix): sqlite3.c
	@$(CC) $(CFLAGS) $(IncludePath) -MG -MP -MT$(IntermediateDirectory)/sqlite3.c$(ObjectSuffix) -MF$(IntermediateDirectory)/sqlite3.c$(DependSuffix) -MM "sqlite3.c"

$(IntermediateDirectory)/sqlite3.c$(PreprocessSuffix): sqlite3.c
	@$(CC) $(CFLAGS) $(IncludePath) $(PreprocessOnlySwitch) $(OutputSwitch) $(IntermediateDirectory)/sqlite3.c$(PreprocessSuffix) "sqlite3.c"


-include $(IntermediateDirectory)/*$(DependSuffix)
##
## Clean
##
clean:
	$(RM) -r ./Release/


