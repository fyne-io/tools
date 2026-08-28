#!/usr/bin/env bash

WORKDIR=`pwd`/work
mkdir -p $WORKDIR

JAVAC=/usr/bin/javac
D8=/opt/android-sdk/build-tools/37.0.0/d8
ANDROID_JAR=/opt/android-sdk/platforms/android-36/android.jar
GO_NATIVE_ACTIVITY=../../../../../fyne/internal/driver/mobile/app/GoNativeActivity.java
NOTIFICATION_RECEIVER=../../../../../fyne/internal/driver/mobile/app/FyneNotificationReceiver.java
BITMAP_SURFACE_PROVIDER=../../../../../fyne/internal/driver/mobile/app/BitmapSurfaceProvider.java

$JAVAC \
	-source 1.8 \
	-target 1.8 \
	-bootclasspath $ANDROID_JAR \
	-cp './d8_libs/*':'./javac_libs/*' \
	-d $WORKDIR/work \
	$GO_NATIVE_ACTIVITY \
	$NOTIFICATION_RECEIVER \
	$BITMAP_SURFACE_PROVIDER

$D8 \
	--min-api 26 \
	--output $WORKDIR \
	--lib $ANDROID_JAR ./d8_libs/*  \
	$WORKDIR/work/org/golang/app/GoNativeActivity*  \
	$WORKDIR/work/org/golang/app/FyneNotificationReceiver*  \
	$WORKDIR/work/org/golang/app/BitmapSurfaceProvider* 

cp $WORKDIR/classes.dex .
go run ./gendex

cd ../.. && go build && cd internal/mobile && cp dex.go ../../../../../fyne/cmd/fyne/internal/mobile/dex.go
