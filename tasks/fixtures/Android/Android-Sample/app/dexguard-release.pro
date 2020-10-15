# DexGuard's default settings are fine for this sample application.

# Display some more statistics about the processed code.
# -verbose

# The Android manifest already refers to the mutlidex support application.
# We now enable multidexing. DexGuard automatically figures out how to
# split the classes.dex file, if the code exceeds the infamous 65K method ID
# limit of the Dalvik file format. The Android runtime then glues the parts
# back together, if necessary with the help of the  multidex compatibility
# library, which you can add to the project.
# This option works in all build systems: Gradle, Ant, Eclipse, Maven,...

# -multidex

-verbose
-dontoptimize

-dontwarn com.newrelic.agent.android.instrumentation.**
-dontwarn android.support.v4.app.**
-dontwarn com.google.gson.**
-dontwarn com.squareup.**

# -libraryjars app/libs
# -dontnote
# -keepattributes *Annotation*