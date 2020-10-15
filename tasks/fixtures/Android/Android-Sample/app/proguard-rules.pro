#updated proguard config 6/25/2014 will use on release v1.5.1
-optimizationpasses 5
-dontusemixedcaseclassnames
-dontskipnonpubliclibraryclasses
-dontpreverify
-verbose
-optimizations !code/simplification/arithmetic,!field/*,!class/merging/*

-printseeds seeds.txt
-printusage unused.txt
-printmapping mapping.txt


################
# keep class names #
################

-keep class com.newrelic.** { *; }
-keep class **$$ModuleAdapter
-keep class **$$InjectAdapter
-keep class **$$StaticInjection
-keep public class com.android.vending.licensing.ILicensingService
#-keep class com.squareup.okhttp.** { *; }
#-keep interface com.squareup.okhttp.** { *; }
#-keep interface okhttp3.** { *; }
-keep class okhttp3.** { *; }
#-keep class okio.** { *; }
-keep class retrofit.** { *; }

-keepclasseswithmembers class * {
    @retrofit.http.* <methods>;
}
-keep class sun.misc.Unsafe { *; }

-keepclassmembers class ** {
    public void onEvent*(**);
}

-keepclasseswithmembernames class * {
    native <methods>;
}

-keepclasseswithmembers class * {
    public <init>(android.content.Context, android.util.AttributeSet);
}

-keepclasseswithmembers class * {
    public <init>(android.content.Context, android.util.AttributeSet, int);
}


-keepclassmembers enum * {
    public static **[] values();
    public static ** valueOf(java.lang.String);
}

-keep class * implements android.os.Parcelable {
  public static final android.os.Parcelable$Creator *;
}

-keepclassmembers,allowobfuscation class * {
    @javax.inject.* *;
    @dagger.* *;
    <init>();
}

##############
# keep attributes #
##############

-renamesourcefileattribute NRSourceFile

-keepattributes *Annotation*
-keepattributes *Signature*
-keepattributes Exceptions, InnerClasses, Annotation, Signature, LineNumberTable, SourceFile, NRSourceFile

##############
#   keep names   #
##############

-keepnames !abstract class adapter.*
-keepnames !abstract class application.*
-keepnames !abstract class dao.*
-keepnames !abstract class event.*
-keepnames !abstract class fragment.*
-keepnames !abstract class model.*
-keepnames !abstract class module.*
-keepnames !abstract class preference.*
-keepnames !abstract class provider.*
-keepnames !abstract class receiver.*
-keepnames !abstract class rest.*
-keepnames !abstract class service.*
-keepnames !abstract class util.*
-keepnames !abstract class view.*
-keepnames class dagger.Lazy

############
#  dont warn  #
############
-dontwarn com.newrelic.**
-dontwarn rx.**
-dontwarn com.squareup.**
-dontwarn org.mockito.**
-dontwarn sun.reflect.**
-dontwarn android.test.**
-dontwarn com.google.appengine.*
-dontwarn com.jakewharton.*
#-dontwarn com.android.support.*
-dontwarn com.google.common.**
-dontwarn org.fest.**
-dontwarn butterknife.**
#-dontwarn android.support.v8.**
#-dontwarn android.support.v4.**
#-dontwarn android.support.v7.**
#-dontwarn android.support.v17.**
#-dontwarn android.support.v13.**
-dontwarn dagger.internal.**
-dontwarn java.nio.file.*
-dontwarn org.codehaus.mojo.animal_sniffer.IgnoreJRERequirement
-dontwarn com.squareup.okhttp.**
-dontwarn rx.**
-dontwarn retrofit.**
-dontwarn retrofit2.Platform$Java8
-dontwarn okhttp3.**
-dontwarn okio.**

